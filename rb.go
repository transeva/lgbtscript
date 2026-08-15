package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"github.com/gorilla/websocket"
)

// ============================================================
// ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ
// ============================================================

var output io.Writer = os.Stdout
var isFileExecution = false
var currentFilePath = ""
var debugMode = false

var (
	user32DLL         = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent    = user32DLL.NewProc("keybd_event")
	procGetWindowText = user32DLL.NewProc("GetWindowTextW")
	procSetWindowText = user32DLL.NewProc("SetWindowTextW")
	procMessageBox    = user32DLL.NewProc("MessageBoxW")
)

// Константы для MessageBox
const (
	MB_OK               = 0x00000000
	MB_OKCANCEL         = 0x00000001
	MB_ABORTRETRYIGNORE = 0x00000002
	MB_YESNOCANCEL      = 0x00000003
	MB_YESNO            = 0x00000004
	MB_RETRYCANCEL      = 0x00000005

	MB_ICONHAND        = 0x00000010
	MB_ICONQUESTION    = 0x00000020
	MB_ICONEXCLAMATION = 0x00000030
	MB_ICONASTERISK    = 0x00000040
	MB_ICONWARNING     = MB_ICONEXCLAMATION
	MB_ICONERROR       = MB_ICONHAND
	MB_ICONINFORMATION = MB_ICONASTERISK

	IDOK     = 1
	IDCANCEL = 2
	IDABORT  = 3
	IDRETRY  = 4
	IDIGNORE = 5
	IDYES    = 6
	IDNO     = 7
)

func MessageBox(hWnd uintptr, lpText, lpCaption *uint16, uType uint32) int {
	ret, _, _ := procMessageBox.Call(
		hWnd,
		uintptr(unsafe.Pointer(lpText)),
		uintptr(unsafe.Pointer(lpCaption)),
		uintptr(uType),
	)
	return int(ret)
}

// ============================================================
// СТРУКТУРЫ ДЛЯ ЧАТА
// ============================================================

type ChatServer struct {
	Name        string
	Clients     map[*websocket.Conn]*ChatClient
	Broadcast   chan ChatMessage
	Register    chan *ChatClient
	Unregister  chan *websocket.Conn
	Mu          sync.RWMutex
	IsActive    bool
	Messages    []ChatMessage
	MaxMessages int
}

type ChatClient struct {
	Conn     *websocket.Conn
	Username string
	Server   *ChatServer
	Room     string
}

type ChatMessage struct {
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room"`
	Type      string    `json:"type"`
}

var (
	chatServers = make(map[string]*ChatServer)
	chatMu      sync.RWMutex
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// ============================================================
// НАСТРОЙКА КОНСОЛИ
// ============================================================

func setupConsole() {
	if os.Getenv("OS") == "Windows_NT" {
		fmt.Fprint(output, "\xef\xbb\xbf")
	}
}

// ============================================================
// ТОКЕНЫ И ЛЕКСЕР
// ============================================================

const (
	TOKEN_NUMBER = iota
	TOKEN_FLOAT
	TOKEN_IDENTIFIER
	TOKEN_KEYWORD
	TOKEN_OPERATOR
	TOKEN_STRING
	TOKEN_EOF
	TOKEN_COMMENT
	TOKEN_INCLUDE
	TOKEN_ARRAY_OPEN
	TOKEN_ARRAY_CLOSE
)

type Token struct {
	Type  int
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	input    string
	pos      int
	line     int
	col      int
	tokens   []Token
	keywords map[string]bool
	err      error
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
		keywords: map[string]bool{
			"lesbian": true, "gay": true, "queer": true, "nonbinary": true, "gender": true,
			"comingout": true, "cis": true, "nocis": true,
			"true": true, "false": true,
			"help": true, "orientation": true,
			"rainbow": true,
			"return": true,
			"hetero": true, "homo": true,
			"export": true,
			"pride": true,
			"sex": true,
			"asexual": true,
		},
		err: nil,
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == ' ' || ch == '\t' {
			l.pos++
			l.col++
			continue
		}

		if ch == '\n' {
			l.pos++
			l.line++
			l.col = 1
			continue
		}

		if ch == '\r' {
			l.pos++
			if l.pos < len(l.input) && l.input[l.pos] == '\n' {
				l.pos++
			}
			l.line++
			l.col = 1
			continue
		}

		if ch == '#' && l.pos+1 < len(l.input) {
			next := l.input[l.pos+1]
			if next >= 'a' && next <= 'z' {
				start := l.pos
				for l.pos < len(l.input) && ((l.input[l.pos] >= 'a' && l.input[l.pos] <= 'z') || l.input[l.pos] == '#') {
					l.pos++
				}
				directive := l.input[start:l.pos]
				if directive == "#inclusive" {
					for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
						l.pos++
						l.col++
					}
					if l.pos < len(l.input) && l.input[l.pos] == '"' {
						l.pos++
						startFile := l.pos
						for l.pos < len(l.input) && l.input[l.pos] != '"' {
							l.pos++
						}
						filename := l.input[startFile:l.pos]
						if l.pos < len(l.input) && l.input[l.pos] == '"' {
							l.pos++
						}
						l.tokens = append(l.tokens, Token{TOKEN_INCLUDE, filename, l.line, l.col})
						for l.pos < len(l.input) && l.input[l.pos] != '\n' {
							l.pos++
						}
						continue
					}
				} else {
					l.pos = start + 1
					l.col++
					continue
				}
			}
		}

		if ch == '@' {
			l.tokenizeComment()
			continue
		}

		if ch == '[' {
			l.pos++
			l.col++
			l.tokens = append(l.tokens, Token{TOKEN_ARRAY_OPEN, "[", l.line, l.col})
			continue
		}

		if ch == ']' {
			l.pos++
			l.col++
			l.tokens = append(l.tokens, Token{TOKEN_ARRAY_CLOSE, "]", l.line, l.col})
			continue
		}

		if ch == '-' && l.pos+1 < len(l.input) {
			next := l.input[l.pos+1]
			if (next >= '0' && next <= '9') || next == '.' {
				l.pos++
				l.col++
				startPos := l.pos - 1
				startLine := l.line
				startCol := l.col - 1
				hasDot := false
				hasDigit := false

				for l.pos < len(l.input) {
					ch2 := l.input[l.pos]
					if ch2 >= '0' && ch2 <= '9' {
						hasDigit = true
						l.pos++
						l.col++
					} else if ch2 == '.' && !hasDot {
						hasDot = true
						l.pos++
						l.col++
					} else {
						break
					}
				}

				if hasDigit {
					value := l.input[startPos:l.pos]
					if hasDot {
						l.tokens = append(l.tokens, Token{TOKEN_FLOAT, value, startLine, startCol})
					} else {
						l.tokens = append(l.tokens, Token{TOKEN_NUMBER, value, startLine, startCol})
					}
					continue
				} else {
					l.pos = startPos + 1
					l.col = startCol + 1
				}
			}
		}

		if (ch >= '0' && ch <= '9') || (ch == '.' && l.pos+1 < len(l.input) && l.input[l.pos+1] >= '0' && l.input[l.pos+1] <= '9') {
			l.tokenizeNumber()
			continue
		}

		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
			l.tokenizeIdentifier()
			continue
		}

		if ch == '"' {
			l.tokenizeString()
			continue
		}

		if err := l.tokenizeOperator(); err != nil {
			return nil, err
		}
	}

	l.tokens = append(l.tokens, Token{TOKEN_EOF, "", l.line, l.col})
	return l.tokens, nil
}

func (l *Lexer) tokenizeComment() {
	startLine := l.line
	startCol := l.col
	l.pos++
	l.col++

	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
		l.pos++
		l.col++
	}

	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '\n' && l.input[l.pos] != '\r' {
		l.pos++
		l.col++
	}

	comment := l.input[start:l.pos]
	comment = strings.TrimSpace(comment)

	l.tokens = append(l.tokens, Token{TOKEN_COMMENT, comment, startLine, startCol})
}

func (l *Lexer) tokenizeNumber() {
	start := l.pos
	startLine := l.line
	startCol := l.col
	hasDot := false

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch >= '0' && ch <= '9' {
			l.pos++
			l.col++
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.pos++
			l.col++
		} else {
			break
		}
	}

	value := l.input[start:l.pos]
	if hasDot {
		l.tokens = append(l.tokens, Token{TOKEN_FLOAT, value, startLine, startCol})
	} else {
		l.tokens = append(l.tokens, Token{TOKEN_NUMBER, value, startLine, startCol})
	}
}

func (l *Lexer) tokenizeIdentifier() {
	start := l.pos
	startLine := l.line
	startCol := l.col

	for l.pos < len(l.input) &&
		((l.input[l.pos] >= 'a' && l.input[l.pos] <= 'z') ||
			(l.input[l.pos] >= 'A' && l.input[l.pos] <= 'Z') ||
			(l.input[l.pos] >= '0' && l.input[l.pos] <= '9') ||
			l.input[l.pos] == '_') {
		l.pos++
		l.col++
	}

	value := l.input[start:l.pos]
	lowerValue := strings.ToLower(value)
	if l.keywords[lowerValue] {
		l.tokens = append(l.tokens, Token{TOKEN_KEYWORD, lowerValue, startLine, startCol})
	} else {
		l.tokens = append(l.tokens, Token{TOKEN_IDENTIFIER, value, startLine, startCol})
	}
}

func (l *Lexer) tokenizeString() {
	start := l.pos
	startLine := l.line
	startCol := l.col
	l.pos++
	l.col++

	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}

	value := l.input[start+1:l.pos]
	if l.pos < len(l.input) && l.input[l.pos] == '"' {
		l.pos++
		l.col++
	}

	l.tokens = append(l.tokens, Token{TOKEN_STRING, value, startLine, startCol})
}

func (l *Lexer) tokenizeOperator() error {
	ch := l.input[l.pos]
	startLine := l.line
	startCol := l.col
	var value string

	if l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		if (ch == '=' && next == '=') || (ch == '!' && next == '=') ||
			(ch == '<' && next == '=') || (ch == '>' && next == '=') ||
			(ch == '&' && next == '&') || (ch == '|' && next == '|') ||
			(ch == '+' && next == '=') || (ch == '-' && next == '=') ||
			(ch == '*' && next == '=') || (ch == '/' && next == '=') ||
			(ch == '%' && next == '=') {
			value = string(ch) + string(next)
			l.pos += 2
			l.col += 2
			l.tokens = append(l.tokens, Token{TOKEN_OPERATOR, value, startLine, startCol})
			return nil
		}
		if (ch == '+' && next == '+') || (ch == '-' && next == '-') {
			value = string(ch) + string(next)
			l.pos += 2
			l.col += 2
			l.tokens = append(l.tokens, Token{TOKEN_OPERATOR, value, startLine, startCol})
			return nil
		}
	}

	operators := "=+-*/%<>=!(){},;."
	if strings.ContainsRune(operators, rune(ch)) {
		value = string(ch)
		l.pos++
		l.col++
		l.tokens = append(l.tokens, Token{TOKEN_OPERATOR, value, startLine, startCol})
		return nil
	}
	return fmt.Errorf("unknown character '%c' at line %d, column %d", ch, startLine, startCol)
}

// ============================================================
// AST УЗЛЫ
// ============================================================

type Node interface {
	GetLine() int
}

type BaseNode struct {
	Line int
	Col  int
}

func (b BaseNode) GetLine() int { return b.Line }

type NumberNode struct {
	BaseNode
	Value int
}
type FloatNode struct {
	BaseNode
	Value float64
}
type StringNode struct {
	BaseNode
	Value string
}
type BooleanNode struct {
	BaseNode
	Value bool
}
type VariableNode struct {
	BaseNode
	Name string
}
type ArrayNode struct {
	BaseNode
	Elements []Node
}
type ArrayIndexNode struct {
	BaseNode
	Name  string
	Index Node
}
type UnaryNode struct {
	BaseNode
	Op   string
	Expr Node
}

type BinaryOpNode struct {
	BaseNode
	Left  Node
	Op    string
	Right Node
}

type TypedDeclaration struct {
	BaseNode
	Type  string
	Name  string
	Value Node
}

type AssignmentStatement struct {
	BaseNode
	Name  string
	Value Node
}

type AugmentedAssignmentStatement struct {
	BaseNode
	Name  string
	Op    string
	Value Node
}

type ArrayAssignmentStatement struct {
	BaseNode
	Name  string
	Index Node
	Value Node
}

type IncrementDecrementStatement struct {
	BaseNode
	Name     string
	Operator string
	Postfix  bool
}

type PrintStatement struct {
	BaseNode
	Value Node
}

type IfStatement struct {
	BaseNode
	Condition    Node
	ThenBlock    []Node
	ElseIfBlocks []ElseIfBlock
	ElseBlock    []Node
}

type ElseIfBlock struct {
	Condition Node
	Block     []Node
}

type WhileStatement struct {
	BaseNode
	Condition Node
	Body      []Node
}

type SexStatement struct {
	BaseNode
	Init      Node
	Condition Node
	Update    Node
	Body      []Node
}

type HelpStatement struct {
	BaseNode
	Country Node
}

type OrientationStatement struct {
	BaseNode
}

type FunctionDeclaration struct {
	BaseNode
	Name     string
	Params   []string
	Body     []Node
	Exported bool
}

type FunctionCall struct {
	BaseNode
	Name string
	Args []Node
}

type ReturnStatement struct {
	BaseNode
	Value Node
}

type CommentStatement struct {
	BaseNode
	Text string
}

type IncludeStatement struct {
	BaseNode
	Filename string
}

type HeteroHomoStatement struct {
	BaseNode
	HeteroBlock []Node
	HomoBlock   []Node
}

type ExpressionStatement struct {
	BaseNode
	Expr Node
}

type ConstantDeclaration struct {
	BaseNode
	Name  string
	Value Node
}

type Program struct {
	BaseNode
	Statements []Node
}

// ============================================================
// ПАРСЕР - ПОЛНАЯ ВЕРСИЯ
// ============================================================

type Parser struct {
	tokens []Token
	pos    int
	err    error
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0, err: nil}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{TOKEN_EOF, "", 0, 0}
}

func (p *Parser) peekNext() Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return Token{TOKEN_EOF, "", 0, 0}
}

func (p *Parser) next() Token {
	token := p.peek()
	p.pos++
	return token
}

func (p *Parser) expect(tokenType int, value string) (Token, error) {
	token := p.peek()
	if token.Type != tokenType {
		return Token{}, fmt.Errorf("expected token type %d, got %s at line %d", tokenType, token.Value, token.Line)
	}
	if value != "" && token.Value != value {
		return Token{}, fmt.Errorf("expected value %s, got %s at line %d", value, token.Value, token.Line)
	}
	p.pos++
	return token, nil
}

func (p *Parser) getCurrentLine() int {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos].Line
	}
	return 0
}

// ============================================================
// ПАРСЕР - МЕТОДЫ РАЗБОРА ВЫРАЖЕНИЙ
// ============================================================

func (p *Parser) parseExpression() (Node, error) {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() (Node, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().Value == "||" {
		op := p.peek().Value
		p.next()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseLogicalAnd() (Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.peek().Value == "&&" {
		op := p.peek().Value
		p.next()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseEquality() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for {
		token := p.peek()
		if token.Value != "==" && token.Value != "!=" {
			break
		}
		op := token.Value
		p.next()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for {
		token := p.peek()
		if token.Value != "<" && token.Value != ">" && token.Value != "<=" && token.Value != ">=" {
			break
		}
		op := token.Value
		p.next()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseAdditive() (Node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for {
		token := p.peek()
		if token.Value != "+" && token.Value != "-" {
			break
		}
		op := token.Value
		p.next()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseMultiplicative() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		token := p.peek()
		if token.Value != "*" && token.Value != "/" && token.Value != "%" {
			break
		}
		op := token.Value
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{
			BaseNode: BaseNode{Line: left.GetLine(), Col: 0},
			Left:     left,
			Op:       op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	token := p.peek()
	if token.Value == "-" || token.Value == "!" {
		op := token.Value
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{
			BaseNode: BaseNode{Line: token.Line, Col: token.Col},
			Op:       op,
			Expr:     expr,
		}, nil
	}

	if token.Value == "++" || token.Value == "--" {
		op := token.Value
		p.next()
		if p.peek().Type != TOKEN_IDENTIFIER {
			return nil, fmt.Errorf("expected identifier after %s", op)
		}
		name := p.peek().Value
		p.next()
		return &IncrementDecrementStatement{
			BaseNode: BaseNode{Line: token.Line, Col: token.Col},
			Name:     name,
			Operator: op,
			Postfix:  false,
		}, nil
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col

	switch token.Type {
	case TOKEN_NUMBER:
		val, err := strconv.Atoi(token.Value)
		if err != nil {
			return nil, err
		}
		p.next()
		return &NumberNode{BaseNode: BaseNode{Line: line, Col: col}, Value: val}, nil

	case TOKEN_FLOAT:
		val, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, err
		}
		p.next()
		return &FloatNode{BaseNode: BaseNode{Line: line, Col: col}, Value: val}, nil

	case TOKEN_STRING:
		p.next()
		return &StringNode{BaseNode: BaseNode{Line: line, Col: col}, Value: token.Value}, nil

	case TOKEN_KEYWORD:
		if token.Value == "true" {
			p.next()
			return &BooleanNode{BaseNode: BaseNode{Line: line, Col: col}, Value: true}, nil
		}
		if token.Value == "false" {
			p.next()
			return &BooleanNode{BaseNode: BaseNode{Line: line, Col: col}, Value: false}, nil
		}

	case TOKEN_IDENTIFIER:
		nextToken := p.peekNext()
		if nextToken.Value == "(" {
			return p.parseFunctionCall(token.Value)
		}
		if nextToken.Value == "[" {
			p.next()
			p.next()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(TOKEN_ARRAY_CLOSE, "]")
			if err != nil {
				return nil, err
			}
			return &ArrayIndexNode{
				BaseNode: BaseNode{Line: line, Col: col},
				Name:     token.Value,
				Index:    index,
			}, nil
		}
		p.next()
		return &VariableNode{BaseNode: BaseNode{Line: line, Col: col}, Name: token.Value}, nil

	case TOKEN_ARRAY_OPEN:
		p.next()
		var elements []Node
		if p.peek().Value != "]" {
			for {
				elem, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, elem)
				if p.peek().Value == "," {
					p.next()
					continue
				}
				break
			}
		}
		_, err := p.expect(TOKEN_ARRAY_CLOSE, "]")
		if err != nil {
			return nil, err
		}
		return &ArrayNode{
			BaseNode: BaseNode{Line: line, Col: col},
			Elements: elements,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	return nil, fmt.Errorf("unexpected token: %s", token.Value)
}

// ============================================================
// ПАРСЕР - РАЗБОР СТАТЕМЕНТОВ
// ============================================================

func (p *Parser) parseStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col

	if token.Type == TOKEN_INCLUDE {
		filename := token.Value
		p.next()
		return &IncludeStatement{BaseNode: BaseNode{Line: line, Col: col}, Filename: filename}, nil
	}

	if token.Type == TOKEN_COMMENT {
		comment := &CommentStatement{BaseNode: BaseNode{Line: line, Col: col}, Text: token.Value}
		p.next()
		return comment, nil
	}

	if token.Type == TOKEN_KEYWORD {
		keyword := token.Value
		switch keyword {
		case "lesbian", "gay", "queer", "nonbinary", "gender":
			return p.parseTypedDeclaration()
		case "comingout":
			return p.parsePrintStatement()
		case "cis":
			return p.parseIfStatement()
		case "pride":
			return p.parseWhileStatement()
		case "sex":
			return p.parseSexStatement()
		case "help":
			return p.parseHelpStatement()
		case "orientation":
			return p.parseOrientationStatement()
		case "rainbow":
			return p.parseFunctionDeclaration()
		case "return":
			return p.parseReturnStatement()
		case "hetero":
			return p.parseHeteroHomoStatement()
		case "export":
			return p.parseExportStatement()
		case "asexual":
			return p.parseConstantDeclaration()
		}
	}

	if token.Type == TOKEN_IDENTIFIER {
		nextToken := p.peekNext()

		if nextToken.Value == "++" || nextToken.Value == "--" {
			return p.parseIncrementDecrement(token.Value, false)
		}

		if nextToken.Value == "(" {
			call, err := p.parseFunctionCall(token.Value)
			if err != nil {
				return nil, err
			}
			if p.peek().Value == ";" {
				p.next()
			}
			return call, nil
		}

		if nextToken.Value == "[" {
			return p.parseArrayAssignment(token.Value)
		}

		if nextToken.Value == "=" || nextToken.Value == "+=" || nextToken.Value == "-=" ||
			nextToken.Value == "*=" || nextToken.Value == "/=" || nextToken.Value == "%=" {
			return p.parseAssignment(token.Value)
		}

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.peek().Value == ";" {
			p.next()
		}
		return &ExpressionStatement{BaseNode: BaseNode{Line: line, Col: col}, Expr: expr}, nil
	}

	if token.Type == TOKEN_OPERATOR {
		if token.Value == "++" || token.Value == "--" {
			p.next()
			if p.peek().Type != TOKEN_IDENTIFIER {
				return nil, fmt.Errorf("expected identifier after %s", token.Value)
			}
			name := p.peek().Value
			p.next()
			if p.peek().Value == ";" {
				p.next()
			}
			return &IncrementDecrementStatement{
				BaseNode: BaseNode{Line: token.Line, Col: token.Col},
				Name:     name,
				Operator: token.Value,
				Postfix:  false,
			}, nil
		}
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &ExpressionStatement{BaseNode: BaseNode{Line: line, Col: col}, Expr: expr}, nil
}

func (p *Parser) parseTypedDeclaration() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	typeName := token.Value
	p.next()

	nameToken, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "=")
	if err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &TypedDeclaration{
		BaseNode: BaseNode{Line: line, Col: col},
		Type:     typeName,
		Name:     nameToken.Value,
		Value:    value,
	}, nil
}

func (p *Parser) parsePrintStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &PrintStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Value:    value,
	}, nil
}

func (p *Parser) parseIfStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var thenBlock []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		thenBlock = append(thenBlock, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	var elseIfBlocks []ElseIfBlock
	var elseBlock []Node

	for p.peek().Value == "nocis" {
		p.next()

		_, err := p.expect(TOKEN_OPERATOR, "(")
		if err != nil {
			return nil, err
		}

		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		_, err = p.expect(TOKEN_OPERATOR, ")")
		if err != nil {
			return nil, err
		}

		_, err = p.expect(TOKEN_OPERATOR, "{")
		if err != nil {
			return nil, err
		}
		var block []Node
		for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			block = append(block, stmt)
		}
		_, err = p.expect(TOKEN_OPERATOR, "}")
		if err != nil {
			return nil, err
		}
		elseIfBlocks = append(elseIfBlocks, ElseIfBlock{Condition: cond, Block: block})
	}

	if p.peek().Value == "cis" {
		p.next()
		_, err = p.expect(TOKEN_OPERATOR, "{")
		if err != nil {
			return nil, err
		}
		for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			elseBlock = append(elseBlock, stmt)
		}
		_, err = p.expect(TOKEN_OPERATOR, "}")
		if err != nil {
			return nil, err
		}
	}

	return &IfStatement{
		BaseNode:     BaseNode{Line: line, Col: col},
		Condition:    cond,
		ThenBlock:    thenBlock,
		ElseIfBlocks: elseIfBlocks,
		ElseBlock:    elseBlock,
	}, nil
}

func (p *Parser) parseWhileStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var body []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &WhileStatement{
		BaseNode:  BaseNode{Line: line, Col: col},
		Condition: cond,
		Body:      body,
	}, nil
}

func (p *Parser) parseSexStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	var init Node
	if p.peek().Value != ";" {
		init, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
	}
	if p.peek().Value == ";" {
		p.next()
	}

	var cond Node
	if p.peek().Value != ";" {
		cond, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	if p.peek().Value == ";" {
		p.next()
	}

	var update Node
	if p.peek().Value != ")" {
		update, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var body []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &SexStatement{
		BaseNode:  BaseNode{Line: line, Col: col},
		Init:      init,
		Condition: cond,
		Update:    update,
		Body:      body,
	}, nil
}

func (p *Parser) parseHelpStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	country, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &HelpStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Country:  country,
	}, nil
}

func (p *Parser) parseOrientationStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	if p.peek().Value == ";" {
		p.next()
	}

	return &OrientationStatement{
		BaseNode: BaseNode{Line: line, Col: col},
	}, nil
}

func (p *Parser) parseFunctionDeclaration() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	name, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	var params []string
	if p.peek().Value != ")" {
		for {
			param, err := p.expect(TOKEN_IDENTIFIER, "")
			if err != nil {
				return nil, err
			}
			params = append(params, param.Value)
			if p.peek().Value == "," {
				p.next()
				continue
			}
			break
		}
	}
	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var body []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &FunctionDeclaration{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name.Value,
		Params:   params,
		Body:     body,
		Exported: false,
	}, nil
}

func (p *Parser) parseReturnStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	var value Node
	if p.peek().Value != ";" {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &ReturnStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Value:    value,
	}, nil
}
func (p *Parser) parseHeteroHomoStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next() // пропускаем "hetero"

	_, err := p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var heteroBlock []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		heteroBlock = append(heteroBlock, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_KEYWORD, "homo")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var homoBlock []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		homoBlock = append(homoBlock, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &HeteroHomoStatement{
		BaseNode:    BaseNode{Line: line, Col: col},
		HeteroBlock: heteroBlock,
		HomoBlock:   homoBlock,
	}, nil
}
func (p *Parser) parseExportStatement() (Node, error) {
	token := p.peek()
	line := token.Line

	p.next()

	rainbowToken := p.peek()
	if rainbowToken.Type != TOKEN_KEYWORD || rainbowToken.Value != "rainbow" {
		return nil, fmt.Errorf("expected 'rainbow' after export at line %d", line)
	}
	p.next()

	fn, err := p.parseFunctionDeclaration()
	if err != nil {
		return nil, err
	}

	if decl, ok := fn.(*FunctionDeclaration); ok {
		decl.Exported = true
		return decl, nil
	}

	return nil, fmt.Errorf("expected function declaration after export at line %d", line)
}

func (p *Parser) parseConstantDeclaration() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	name, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "=")
	if err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &ConstantDeclaration{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name.Value,
		Value:    value,
	}, nil
}

func (p *Parser) parseAssignment(name string) (Node, error) {
	line := p.peek().Line
	col := p.peek().Col

	op := p.peekNext().Value
	p.pos += 2

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	if op == "=" {
		return &AssignmentStatement{
			BaseNode: BaseNode{Line: line, Col: col},
			Name:     name,
			Value:    value,
		}, nil
	}

	return &AugmentedAssignmentStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name,
		Op:       op[:len(op)-1],
		Value:    value,
	}, nil
}

func (p *Parser) parseArrayAssignment(name string) (Node, error) {
	line := p.peek().Line
	col := p.peek().Col

	p.next()
	_, err := p.expect(TOKEN_ARRAY_OPEN, "[")
	if err != nil {
		return nil, err
	}

	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_ARRAY_CLOSE, "]")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "=")
	if err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().Value == ";" {
		p.next()
	}

	return &ArrayAssignmentStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name,
		Index:    index,
		Value:    value,
	}, nil
}

func (p *Parser) parseIncrementDecrement(name string, postfix bool) (Node, error) {
	line := p.peek().Line
	col := p.peek().Col

	op := p.peekNext().Value
	p.pos += 2

	if p.peek().Value == ";" {
		p.next()
	}

	return &IncrementDecrementStatement{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name,
		Operator: op,
		Postfix:  postfix,
	}, nil
}

func (p *Parser) parseFunctionCall(name string) (Node, error) {
	line := p.peek().Line
	col := p.peek().Col

	p.next()
	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}

	var args []Node
	if p.peek().Value != ")" {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.peek().Value == "," {
				p.next()
				continue
			}
			break
		}
	}
	_, err = p.expect(TOKEN_OPERATOR, ")")
	if err != nil {
		return nil, err
	}

	return &FunctionCall{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name,
		Args:     args,
	}, nil
}

// ============================================================
// ПАРСЕР - ОСНОВНОЙ МЕТОД
// ============================================================

func (p *Parser) Parse() (*Program, error) {
	var statements []Node

	for p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	return &Program{
		BaseNode:   BaseNode{Line: 0, Col: 0},
		Statements: statements,
	}, nil
}

// ============================================================
// СИСТЕМА ТИПОВ
// ============================================================

type ValueType int

const (
	TypeString ValueType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeArray
	TypeObject
	TypeNull
)

type TypedValue struct {
	Type  ValueType
	Value interface{}
}

func (tv TypedValue) String() string {
	switch tv.Type {
	case TypeString:
		return tv.Value.(string)
	case TypeInt:
		return strconv.Itoa(tv.Value.(int))
	case TypeFloat:
		return strconv.FormatFloat(tv.Value.(float64), 'f', -1, 64)
	case TypeBool:
		return strconv.FormatBool(tv.Value.(bool))
	case TypeArray:
		arr := tv.Value.([]TypedValue)
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = v.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return "null"
	}
}

func NewTypedString(s string) TypedValue {
	return TypedValue{Type: TypeString, Value: s}
}

func NewTypedInt(i int) TypedValue {
	return TypedValue{Type: TypeInt, Value: i}
}

func NewTypedFloat(f float64) TypedValue {
	return TypedValue{Type: TypeFloat, Value: f}
}

func NewTypedBool(b bool) TypedValue {
	return TypedValue{Type: TypeBool, Value: b}
}

func NewTypedArray(arr []TypedValue) TypedValue {
	return TypedValue{Type: TypeArray, Value: arr}
}

// ============================================================
// HATE FILTER - ПОЛНАЯ ВЕРСИЯ С МОРФОЛОГИЕЙ
// ============================================================

type HateFilter struct {
	slurs         map[string][]string
	patterns      []*regexp.Regexp
	falsePositive map[string]bool
	mu            sync.RWMutex
	enabled       bool
	action        string
	logFile       string
	morphCache    map[string][]string
}

func NewHateFilter() *HateFilter {
	hf := &HateFilter{
		slurs:         make(map[string][]string),
		patterns:      []*regexp.Regexp{},
		falsePositive: make(map[string]bool),
		enabled:       true,
		action:        "warn",
		logFile:       "hate_speech_log.txt",
		morphCache:    make(map[string][]string),
	}

	hf.initSlurs()
	hf.compilePatterns()
	return hf
}

func (hf *HateFilter) generateMorphVariants(word string) []string {
	variants := []string{word}
	lower := strings.ToLower(word)

	variants = append(variants, lower)

	// Русские суффиксы
	ruSuffixes := []string{
		"", "а", "у", "ом", "е", "ы", "ов", "ам", "ами", "ах",
		"ка", "ку", "кой", "кою", "ке", "ки", "к",
		"чик", "чица", "чику", "чиком", "чике",
		"ище", "ища", "ищу", "ищем", "ище",
		"онок", "онка", "онку", "онком", "онке",
		"енок", "енка", "енку", "енком", "енке",
		"ик", "ики", "ок", "очек", "ечек", "ичек",
		"улька", "улечка", "ушечка", "ишечка",
		"яра", "яры", "юга", "юги",
	}

	// Английские суффиксы
	enSuffixes := []string{
		"", "s", "es", "ed", "ing", "er", "est",
		"ly", "ness", "less", "ful", "ous", "ive",
		"able", "ible", "tion", "sion", "ment", "ity",
		"ism", "ist", "ize", "isation", "ization",
		"y", "ie", "ies", "ey", "o", "oh",
	}

	allSuffixes := append(ruSuffixes, enSuffixes...)

	for _, suffix := range allSuffixes {
		if suffix != "" {
			variants = append(variants, lower+suffix)
			variants = append(variants, word+suffix)
		}
	}

	// Префиксы
	prefixes := []string{
		"", "анти", "проти", "контр", "супер", "гипер",
		"мега", "ультра", "экстра", "транс", "интер",
	}

	for _, prefix := range prefixes {
		if prefix != "" {
			variants = append(variants, prefix+lower)
			variants = append(variants, prefix+word)
		}
	}

	uniqueMap := make(map[string]bool)
	var uniqueVariants []string
	for _, v := range variants {
		if !uniqueMap[v] {
			uniqueMap[v] = true
			uniqueVariants = append(uniqueVariants, v)
		}
	}

	return uniqueVariants
}

func (hf *HateFilter) initSlurs() {
	// RUSSIAN
	hf.slurs["ru"] = []string{
		"пидор", "пидорас", "педрила", "петух", "гомик", "гомосек",
		"гомосеки", "гомосексуалист", "голубой", "голубые", "гейский",
		"гей-пропаганда", "нетрадиционный", "нетрадиционные",
		"извращенец", "извращенка", "извращенцы", "содомит",
		"мужеложец", "мужеложники", "педераст", "педерастия",
		"гей-свадьба", "однополый", "однополые", "гомосятина",
		"трансвестит", "трансгендер", "транс-извращенец",
		"переодетый", "переодетая", "мужик в юбке",
		"баба с яйцами", "женщина с членом", "мужик с сиськами",
		"транс-пропаганда", "транс-агитация", "транс-маньяк",
		"бисексуал", "би-пропаганда", "би-извращенец",
		"двуличный", "двуполый", "лгбт-пропаганда", "содомия",
		"половая извращение", "сексуальное извращение",
		"разврат", "развратник", "педофил", "педофилия",
		"толерастия", "толераст", "либераст", "гомосек",
		"транс", "трансы", "трансик", "трансики",
		"транс-активист", "транс-активисты", "транс-сообщество",
		"транс-человек", "транс-люди", "транс-женщина", "транс-мужчина",
	}

	// ENGLISH
	hf.slurs["en"] = []string{
		"fag", "faggot", "faggy", "faggotry", "homo", "homosexual",
		"homos", "homoerotic", "queer", "dyke", "dykes", "lezzie",
		"lezbo", "butch", "femme", "twink", "gayboy", "gaylord",
		"gaywad", "gayfag", "sodomite", "sodomy", "bugger",
		"buggery", "poof", "poofter", "puff", "puffer", "nancy",
		"nancyboy", "sissy", "pansy", "cocksucker", "cocksucking",
		"assbandit", "fruit", "fruitcake", "fruity", "fairy",
		"queen", "drag queen", "queerbaiter", "tranny", "trannies",
		"transexual", "transvestite", "shemale", "he-she", "traps",
		"trap", "ladyboy", "chick with dick", "dickgirl", "cuntboy",
		"transgenderism", "gender ideology", "gender confusion",
		"trans propaganda",
		"gaystapo", "groomer", "grooming",
		"degenerate", "degeneracy", "perversion", "deviant",
		"abomination", "unnatural", "satanic",
		"trans", "transes", "transy",
	}

	// SPANISH
	hf.slurs["es"] = []string{
		"maricón", "marica", "maricon", "maricona", "joto", "jota",
		"puto", "puta", "homosexual", "lesbiana", "travesti",
		"travestido", "transformista", "loca", "locas", "afeminado",
		"afeminada", "bollera", "bollo", "tortillera", "camionera",
		"camionero", "machorra", "sodomita", "homofobia", "transfobia",
		"ideología de género", "teoría de género", "agenda gay",
	}

	// FRENCH
	hf.slurs["fr"] = []string{
		"pédé", "pédale", "tapette", "tante", "folle", "homosexuel",
		"homosexuelle", "lesbienne", "travesti", "transsexuel",
		"travelo", "trave", "fif", "efféminé", "efféminée",
		"sodomite", "sodomie", "tarlouze", "petite pédale",
		"homosexualité", "idéologie du genre", "propagande gay",
		"lobby gay", "pédophilie", "pédophile", "pervers", "perversion",
	}

	// GERMAN
	hf.slurs["de"] = []string{
		"Schwuler", "Schwule", "Schwuchtel", "Tunte", "Homosexuell",
		"Homosexuelle", "Lesbe", "Transvestit", "Transsexual",
		"Transgender", "Transe", "Tranny", "Travestie", "Weichei",
		"Sodomit", "Sodomie", "Homo", "Homolobby", "Genderideologie",
		"Genderwahn", "Regenbogenmafia", "Kinderschänder", "Pädophiler",
		"widernatürlich", "pervers",
	}

	// ITALIAN
	hf.slurs["it"] = []string{
		"finocchio", "finocchia", "ricchione", "culattone",
		"omosessuale", "lesbica", "transessuale", "travestito",
		"transgender", "femminiello", "effeminato", "sodomita",
		"sodomia", "lobby gay", "ideologia gender", "pervertito",
		"deviato", "anormale", "contro natura", "peccato", "abominio",
	}

	// PORTUGUESE
	hf.slurs["pt"] = []string{
		"viado", "veado", "bicha", "boiola", "boia", "homossexual",
		"lésbica", "sapatão", "travesti", "transexual", "transgênero",
		"transformista", "afeminado", "sodomita", "sodomia",
		"ideologia de gênero", "propaganda gay", "lobby gay",
		"pervertido", "desviado", "anormal", "contra a natureza",
		"pecado", "abominação", "groomer", "pedófilo",
	}

	// DUTCH
	hf.slurs["nl"] = []string{
		"flikker", "flikkers", "nicht", "nichten", "homo",
		"homoseksueel", "lesbienne", "pot", "travestiet",
		"transseksueel", "transgender", "sodomiet", "sodomie",
		"genderideologie", "regenboogmafia", "pedofiel",
		"pervers", "tegennatuurlijk", "zonde",
	}

	// POLISH
	hf.slurs["pl"] = []string{
		"pedał", "pedały", "ciapa", "ciota", "homoseksualista",
		"lesbijka", "gej", "transwestyta", "transseksualista",
		"sodomita", "sodomia", "ideologia gender", "teoria gender",
		"propaganda LGBT", "lobby LGBT", "pedofil", "pedofilia",
		"zboczeniec", "dewiacja", "perwersja",
	}

	// UKRAINIAN
	hf.slurs["uk"] = []string{
		"підор", "підорас", "педер", "гомосек", "гомік", "гей",
		"гомосексуал", "лесбі", "трансвестит", "транссексуал",
		"содомія", "содоміт", "мужеложець", "ідеологія гендеру",
		"гендерна ідеологія",
		"педофіл", "педофілія", "збоченець", "девіація",
	}

	// ARABIC
	hf.slurs["ar"] = []string{
		"لوطي", "لوطية", "شاذ", "شاذة", "مثلي", "مثلية",
		"متحول جنسي", "خنثى", "خنيث", "مخنث", "فاحشة", "فجور",
		"زنا", "شذوذ", "انحراف", "مرض نفسي", "ضد الفطرة",
		"ضد الطبيعة", "ترويج الشذوذ", "دعاية مثلية",
	}

	// HEBREW
	hf.slurs["he"] = []string{
		"הומו", "הומוסקסואל", "לסבית", "טרנס", "טרנסג'נדר",
		"טרנסקסואל", "סדום", "סדומיה", "מעשה סדום", "תועבה",
		"תועבות", "סטייה מינית", "אג'נדה הומוסקסואלית",
		"תעמולה הומוסקסואלית", "חריג", "לא נורמלי", "מגונה",
	}

	// TURKISH
	hf.slurs["tr"] = []string{
		"ibne", "ibneler", "göt", "götveren", "eşcinsel",
		"lezbiyen", "homoseksüel", "travesti", "transseksüel",
		"transgender", "sodomi", "sodomist", "doğaya aykırı",
		"cinsiyet ideolojisi", "toplumsal cinsiyet",
		"gay lobisi", "pedofil", "pedofili", "sapık", "anormal",
		"günah", "ayıp", "utanç verici",
	}

	// JAPANESE
	hf.slurs["ja"] = []string{
		"おかま", "おかま野郎", "ホモ", "ゲイ", "レズ", "レズビアン",
		"バイセクシュアル", "オカマ", "オナベ", "ニューハーフ",
		"トランスジェンダー", "性同一性障害", "ゲイのプロパガンダ",
		"異常性欲", "性的倒錯", "変態",
		"反自然的", "不道徳", "罪", "ペドフィリア", "小児性愛",
	}

	// CHINESE
	hf.slurs["zh"] = []string{
		"同性恋", "同性恋者", "基佬", "玻璃", "拉拉", "女同",
		"男同", "同志", "变性人", "人妖", "阴阳人",
		"同性恋宣传", "同性恋议程", "性变态",
		"性倒错", "反常", "不道德", "伤风败俗", "罪恶",
		"恋童癖", "性侵", "猥亵", "性别意识形态",
	}

	// KOREAN
	hf.slurs["ko"] = []string{
		"호모", "게이", "레즈", "동성애자", "동성애",
		"트랜스젠더", "성전환", "변태", "성도착", "반자연적",
		"동성애 선전", "포르노", "음란", "추행",
		"아동 성학대", "페도필리아", "타락", "부도덕", "죄악",
	}

	// HINDI
	hf.slurs["hi"] = []string{
		"समलैंगिक", "समलिंगी", "गे", "लेस्बियन", "किन्नर",
		"हिजड़ा", "ट्रांसजेंडर", "अप्राकृतिक", "अनैतिक", "पाप",
		"यौन विचलन", "यौन विकृति", "बाल यौन शोषण", "पीडोफिलिया",
		"एलजीबीटी एजेंडा", "एलजीबीटी प्रचार", "भ्रष्ट", "विकृत",
	}
}

func (hf *HateFilter) compilePatterns() {
	hf.mu.Lock()
	defer hf.mu.Unlock()

	hf.patterns = []*regexp.Regexp{}
	hf.morphCache = make(map[string][]string)

	allSlurs := []string{}
	for _, slurs := range hf.slurs {
		allSlurs = append(allSlurs, slurs...)
	}

	for _, slur := range allSlurs {
		morphVariants := hf.generateMorphVariants(slur)
		hf.morphCache[slur] = morphVariants

		for _, variant := range morphVariants {
			pattern := regexp.QuoteMeta(variant)
			patterns := []string{
				pattern,
				`\b` + pattern + `\b`,
				pattern + `\s*`,
				`\s*` + pattern,
				`(?i)` + pattern,
				`(?i)\b` + pattern + `\b`,
			}

			for _, p := range patterns {
				re, err := regexp.Compile(p)
				if err == nil {
					hf.patterns = append(hf.patterns, re)
				}
			}
		}
	}
}

func (hf *HateFilter) Check(text string) (bool, []string, []string) {
	if strings.EqualFold(strings.TrimSpace(text), "лгбт") || strings.EqualFold(strings.TrimSpace(text), "lgbt") {
		return false, nil, nil
	}
	hf.mu.RLock()
	defer hf.mu.RUnlock()

	if !hf.enabled {
		return false, nil, nil
	}

	detected := []string{}
	languages := []string{}
	lowerText := strings.ToLower(text)

	for _, re := range hf.patterns {
		if re.MatchString(text) {
			detected = append(detected, re.String())
		}
	}

	words := strings.Fields(lowerText)
	for _, word := range words {
		cleanWord := strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})

		if cleanWord == "" {
			continue
		}

		for lang, slurs := range hf.slurs {
			for _, slur := range slurs {
				variants, ok := hf.morphCache[slur]
				if !ok {
					variants = hf.generateMorphVariants(slur)
					hf.morphCache[slur] = variants
				}

				for _, variant := range variants {
					if strings.Contains(cleanWord, variant) ||
						strings.Contains(variant, cleanWord) ||
						strings.Contains(strings.ToLower(cleanWord), strings.ToLower(variant)) {
						if !contains(languages, lang) {
							languages = append(languages, lang)
						}
						if !contains(detected, slur) {
							detected = append(detected, slur)
						}
						break
					}
				}
			}
		}
	}

	phrases := []string{
		"мужик в юбке", "баба с яйцами", "женщина с членом",
		"lgbt agenda", "gender ideology", "rainbow mafia",
		"транс-активист", "транс-сообщество", "транс-человек",
		"транс-женщина", "транс-мужчина", "транс-люди",
	}

	for _, phrase := range phrases {
		if strings.Contains(lowerText, strings.ToLower(phrase)) {
			detected = append(detected, phrase)
			for lang := range hf.slurs {
				if strings.Contains(strings.ToLower(phrase), strings.ToLower(lang)) ||
					strings.Contains(lowerText, strings.ToLower(lang)) {
					if !contains(languages, lang) {
						languages = append(languages, lang)
					}
				}
			}
		}
	}

	return len(detected) > 0, detected, languages
}

func (hf *HateFilter) FilterText(text string) string {
	hf.mu.RLock()
	defer hf.mu.RUnlock()

	result := text
	for _, re := range hf.patterns {
		if re.MatchString(result) {
			result = re.ReplaceAllString(result, "🏳️‍🌈[заблокировано]🏳️‍🌈")
		}
	}

	words := strings.Fields(result)
	for i, word := range words {
		cleanWord := strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})

		if cleanWord == "" {
			continue
		}

		for _, slurs := range hf.slurs {
			for _, slur := range slurs {
				variants, ok := hf.morphCache[slur]
				if !ok {
					variants = hf.generateMorphVariants(slur)
					hf.morphCache[slur] = variants
				}

				for _, variant := range variants {
					if strings.Contains(strings.ToLower(cleanWord), strings.ToLower(variant)) {
						words[i] = strings.ReplaceAll(words[i], cleanWord, "🏳️‍🌈[заблокировано]🏳️‍🌈")
						break
					}
				}
			}
		}
	}

	return strings.Join(words, " ")
}

func (hf *HateFilter) LogHateSpeech(text string, detected []string, languages []string) {
	if hf.logFile == "" {
		return
	}

	f, err := os.OpenFile(hf.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] Найдена ненависть: %s\n", timestamp, text)
	logEntry += fmt.Sprintf("  Обнаружено: %v\n", detected)
	logEntry += fmt.Sprintf("  Языки: %v\n", languages)
	logEntry += "------------------------------------------------\n"

	f.WriteString(logEntry)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================
// ШИФРОВАНИЕ ДЛЯ ДНЕВНИКА
// ============================================================

type Encryption struct {
	key []byte
}

func NewEncryption(key string) *Encryption {
	hashedKey := make([]byte, 32)
	for i := 0; i < 32 && i < len(key); i++ {
		hashedKey[i] = key[i]
	}
	return &Encryption{key: hashedKey}
}

func (e *Encryption) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryption) Decrypt(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ============================================================
// ИНТЕРПРЕТАТОР - ПОЛНАЯ ВЕРСИЯ
// ============================================================

type callFrame struct {
	vars      map[string]TypedValue
	types     map[string]string
	constants map[string]bool
}

type Sandbox struct {
	allowedPaths   []string
	blockedPaths   []string
	maxFileSize    int64
	allowedDomains []string
	blockedDomains []string
}

func NewSandbox() *Sandbox {
	return &Sandbox{
		allowedPaths:   []string{".", "libs", "libraries"},
		blockedPaths:   []string{"/etc", "/proc", "/sys", "~", "/root", "/home"},
		maxFileSize:    10 * 1024 * 1024,
		allowedDomains: []string{},
		blockedDomains: []string{"localhost", "127.0.0.1"},
	}
}

func (s *Sandbox) CheckFilePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}

	for _, blocked := range s.blockedPaths {
		if strings.Contains(absPath, blocked) {
			return fmt.Errorf("access denied: path contains blocked directory '%s'", blocked)
		}
	}
	return nil
}

func (s *Sandbox) CheckURL(url string) error {
	for _, blocked := range s.blockedDomains {
		if strings.Contains(url, blocked) {
			return fmt.Errorf("access denied: domain '%s' is blocked", blocked)
		}
	}
	return nil
}

type JournalEntry struct {
	Timestamp time.Time
	Text      string
	Mood      string
	Encrypted bool
}

type EventEntry struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Date        string   `json:"date"`
	Country     string   `json:"country"`
	City        string   `json:"city"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	Organizer   string   `json:"organizer,omitempty"`
	Tags        []string `json:"tags"`
}

type BookEntry struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Year        string   `json:"year"`
	Genre       string   `json:"genre"`
	Description string   `json:"description"`
	Themes      []string `json:"themes"`
	AgeGroup    string   `json:"age_group"`
}

type Interpreter struct {
	variables      map[string]TypedValue
	variableTypes  map[string]string
	functions      map[string]*FunctionDeclaration
	exportedFuncs  map[string]*FunctionDeclaration
	callStack      []callFrame
	returnValue    TypedValue
	returnFlag     bool
	mu             sync.RWMutex
	maxRecursion   int
	recursionDepth int
	errorHandler   func(error, int, int)
	sandbox        *Sandbox
	rand           *mathrand.Rand
	hateFilter     *HateFilter
	encryption     *Encryption
	journal        []JournalEntry
	journalMu      sync.RWMutex
	events         []EventEntry
	eventsMu       sync.RWMutex
	quotes         []string
	books          []BookEntry
	flagColors     map[string][]string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:      make(map[string]TypedValue),
		variableTypes:  make(map[string]string),
		functions:      make(map[string]*FunctionDeclaration),
		exportedFuncs:  make(map[string]*FunctionDeclaration),
		callStack:      []callFrame{{vars: make(map[string]TypedValue), types: make(map[string]string), constants: make(map[string]bool)}},
		returnValue:    TypedValue{Type: TypeNull, Value: nil},
		returnFlag:     false,
		maxRecursion:   1000,
		recursionDepth: 0,
		sandbox:        NewSandbox(),
		rand:           mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		hateFilter:     NewHateFilter(),
		encryption:     NewEncryption("LGBTScriptSecretKey2024"),
		journal:        []JournalEntry{},
		quotes: []string{
			"Быть собой — это самый смелый поступок. - Марша П. Джонсон",
			"Любовь — это любовь. - Леди Гага",
			"Каждый имеет право на счастье. - Харви Милк",
			"Мы не должны прятаться. - Байярд Растин",
			"Твоя идентичность — твоя сила. - Лаверн Кокс",
			"Гордость — это сопротивление. - Алан Камминг",
			"Мы все рождены свободными. - Нельсон Мандела",
			"Разнообразие — это красота жизни. - Одри Лорд",
			"Любовь побеждает страх. - Джеймс Болдуин",
			"Ты важен такой, какой ты есть. - Адам Риппон",
		},
		books: []BookEntry{
			{"Гордость и предубеждение", "Джейн Остин", "1813", "Классика", "История о любви и самопринятии", []string{"любовь", "общество", "самоопределение"}, "16+"},
			{"Цвета радуги", "Майкл Генри", "2020", "Детская", "Книга о принятии себя и других", []string{"принятие", "дружба", "разнообразие"}, "6+"},
			{"Моя сестра", "Элис Уокер", "1982", "Драма", "История о сестрах и их пути", []string{"семья", "женщины", "борьба"}, "18+"},
		},
		flagColors: map[string][]string{
			"pride":     {"красный", "оранжевый", "желтый", "зеленый", "синий", "фиолетовый"},
			"trans":     {"голубой", "розовый", "белый", "розовый", "голубой"},
			"lesbian":   {"оранжевый", "белый", "розовый"},
			"gay":       {"голубой", "зеленый", "белый"},
			"bisexual":  {"розовый", "фиолетовый", "синий"},
			"nonbinary": {"желтый", "белый", "фиолетовый", "черный"},
		},
	}
}

func (i *Interpreter) checkHateSpeech(text string) (bool, string) {
	if !i.hateFilter.enabled {
		return false, ""
	}

	hasHate, detected, languages := i.hateFilter.Check(text)
	if hasHate {
		i.hateFilter.LogHateSpeech(text, detected, languages)
		filtered := i.hateFilter.FilterText(text)
		return true, filtered
	}
	return false, text
}

func (i *Interpreter) Evaluate(node Node) (TypedValue, error) {
	if program, ok := node.(*Program); ok {
		return i.evaluateProgram(program)
	}
	return i.evaluateNode(node)
}

func (i *Interpreter) evaluateProgram(program *Program) (TypedValue, error) {
	for _, stmt := range program.Statements {
		_, err := i.evaluateNode(stmt)
		if err != nil {
			return TypedValue{}, err
		}
	}
	return TypedValue{Type: TypeNull, Value: nil}, nil
}

func (i *Interpreter) evaluateNode(node Node) (TypedValue, error) {
	switch n := node.(type) {
	case *IncludeStatement:
		return i.handleInclude(n)
	case *CommentStatement:
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *NumberNode:
		return NewTypedInt(n.Value), nil
	case *FloatNode:
		return NewTypedFloat(n.Value), nil
	case *StringNode:
		return NewTypedString(n.Value), nil
	case *BooleanNode:
		return NewTypedBool(n.Value), nil
	case *VariableNode:
		if val, ok := i.getVar(n.Name); ok {
			return val, nil
		}
		return TypedValue{}, fmt.Errorf("variable not defined: %s", n.Name)
	case *ArrayNode:
		elements := make([]TypedValue, len(n.Elements))
		for idx, elem := range n.Elements {
			val, err := i.evaluateNode(elem)
			if err != nil {
				return TypedValue{}, err
			}
			elements[idx] = val
		}
		return NewTypedArray(elements), nil
	case *ArrayIndexNode:
		arr, ok := i.getVar(n.Name)
		if !ok {
			return TypedValue{}, fmt.Errorf("array not defined: %s", n.Name)
		}
		if arr.Type != TypeArray {
			return TypedValue{}, fmt.Errorf("variable %s is not an array", n.Name)
		}
		arrSlice := arr.Value.([]TypedValue)
		idxVal, err := i.evaluateNode(n.Index)
		if err != nil {
			return TypedValue{}, err
		}
		if idxVal.Type != TypeInt {
			return TypedValue{}, fmt.Errorf("array index must be integer")
		}
		idx := idxVal.Value.(int)
		if idx < 0 || idx >= len(arrSlice) {
			return TypedValue{}, fmt.Errorf("array index out of bounds: %d", idx)
		}
		return arrSlice[idx], nil
	case *UnaryNode:
		val, err := i.evaluateNode(n.Expr)
		if err != nil {
			return TypedValue{}, err
		}
		switch n.Op {
		case "-":
			switch val.Type {
			case TypeInt:
				return NewTypedInt(-val.Value.(int)), nil
			case TypeFloat:
				return NewTypedFloat(-val.Value.(float64)), nil
			default:
				return TypedValue{}, fmt.Errorf("cannot apply unary minus to type %v", val.Type)
			}
		case "!":
			return NewTypedBool(!i.isTruthy(val)), nil
		default:
			return TypedValue{}, fmt.Errorf("unknown unary operator: %s", n.Op)
		}
	case *BinaryOpNode:
		left, err := i.evaluateNode(n.Left)
		if err != nil {
			return TypedValue{}, err
		}
		right, err := i.evaluateNode(n.Right)
		if err != nil {
			return TypedValue{}, err
		}
		return i.evaluateBinaryOp(left, n.Op, right)
	case *TypedDeclaration:
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		if err := i.checkType(n.Type, value); err != nil {
			return TypedValue{}, err
		}
		i.setVar(n.Name, value)
		i.setType(n.Name, n.Type)
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *AssignmentStatement:
		if i.isConstant(n.Name) {
			return TypedValue{}, fmt.Errorf("cannot assign to constant '%s'", n.Name)
		}
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		if typ, ok := i.getType(n.Name); ok {
			if err := i.checkType(typ, value); err != nil {
				return TypedValue{}, err
			}
			i.setVar(n.Name, value)
		} else {
			return TypedValue{}, fmt.Errorf("variable not declared: %s", n.Name)
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *AugmentedAssignmentStatement:
		if i.isConstant(n.Name) {
			return TypedValue{}, fmt.Errorf("cannot assign to constant '%s'", n.Name)
		}
		current, ok := i.getVar(n.Name)
		if !ok {
			return TypedValue{}, fmt.Errorf("variable not defined: %s", n.Name)
		}
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}

		var result TypedValue
		switch n.Op {
		case "+":
			result, err = i.add(current, value)
		case "-":
			result, err = i.subtract(current, value)
		case "*":
			result, err = i.multiply(current, value)
		case "/":
			result, err = i.divide(current, value)
		case "%":
			result, err = i.modulo(current, value)
		default:
			return TypedValue{}, fmt.Errorf("unknown augmented operator: %s", n.Op)
		}
		if err != nil {
			return TypedValue{}, err
		}

		if typ, ok := i.getType(n.Name); ok {
			if err := i.checkType(typ, result); err != nil {
				return TypedValue{}, err
			}
		}
		i.setVar(n.Name, result)
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *IncrementDecrementStatement:
		if i.isConstant(n.Name) {
			return TypedValue{}, fmt.Errorf("cannot modify constant '%s'", n.Name)
		}
		current, ok := i.getVar(n.Name)
		if !ok {
			return TypedValue{}, fmt.Errorf("variable not defined: %s", n.Name)
		}

		var result TypedValue
		switch current.Type {
		case TypeInt:
			val := current.Value.(int)
			if n.Operator == "++" {
				val++
			} else {
				val--
			}
			result = NewTypedInt(val)
		case TypeFloat:
			val := current.Value.(float64)
			if n.Operator == "++" {
				val++
			} else {
				val--
			}
			result = NewTypedFloat(val)
		default:
			return TypedValue{}, fmt.Errorf("cannot increment/decrement type %v", current.Type)
		}

		if n.Postfix {
			oldValue := current
			i.setVar(n.Name, result)
			return oldValue, nil
		}
		i.setVar(n.Name, result)
		return result, nil
	case *ArrayAssignmentStatement:
		if i.isConstant(n.Name) {
			return TypedValue{}, fmt.Errorf("cannot assign to constant array '%s'", n.Name)
		}
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		arr, ok := i.getVar(n.Name)
		if !ok {
			return TypedValue{}, fmt.Errorf("array not defined: %s", n.Name)
		}
		if arr.Type != TypeArray {
			return TypedValue{}, fmt.Errorf("variable %s is not an array", n.Name)
		}
		arrSlice := arr.Value.([]TypedValue)
		idxVal, err := i.evaluateNode(n.Index)
		if err != nil {
			return TypedValue{}, err
		}
		if idxVal.Type != TypeInt {
			return TypedValue{}, fmt.Errorf("array index must be integer")
		}
		idx := idxVal.Value.(int)
		if idx < 0 || idx >= len(arrSlice) {
			return TypedValue{}, fmt.Errorf("array index out of bounds: %d", idx)
		}
		arrSlice[idx] = value
		i.setVar(n.Name, NewTypedArray(arrSlice))
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *PrintStatement:
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}

		text := value.String()
		hasHate, filtered := i.checkHateSpeech(text)
		if hasHate {
			fmt.Fprintln(output, "⚠️ Обнаружена ненависть в выводе!")
			fmt.Fprintln(output, "📝 Отфильтрованный вывод:", filtered)
		} else {
			fmt.Fprintln(output, text)
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *IfStatement:
		cond, err := i.evaluateNode(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}
		if i.isTruthy(cond) {
			for _, stmt := range n.ThenBlock {
				_, err := i.evaluateNode(stmt)
				if err != nil {
					return TypedValue{}, err
				}
			}
			return TypedValue{Type: TypeNull, Value: nil}, nil
		}

		for _, elseIf := range n.ElseIfBlocks {
			cond, err := i.evaluateNode(elseIf.Condition)
			if err != nil {
				return TypedValue{}, err
			}
			if i.isTruthy(cond) {
				for _, stmt := range elseIf.Block {
					_, err := i.evaluateNode(stmt)
					if err != nil {
						return TypedValue{}, err
					}
				}
				return TypedValue{Type: TypeNull, Value: nil}, nil
			}
		}

		for _, stmt := range n.ElseBlock {
			_, err := i.evaluateNode(stmt)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *WhileStatement:
		condVal, err := i.evaluateNode(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}
		for i.isTruthy(condVal) {
			for _, stmt := range n.Body {
				_, err := i.evaluateNode(stmt)
				if err != nil {
					return TypedValue{}, err
				}
				if i.returnFlag {
					return i.returnValue, nil
				}
			}
			condVal, err = i.evaluateNode(n.Condition)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *SexStatement:
		if n.Init != nil {
			_, err := i.evaluateNode(n.Init)
			if err != nil {
				return TypedValue{}, err
			}
		}

		condVal, err := i.evaluateNode(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}

		for i.isTruthy(condVal) {
			for _, stmt := range n.Body {
				_, err := i.evaluateNode(stmt)
				if err != nil {
					return TypedValue{}, err
				}
				if i.returnFlag {
					return i.returnValue, nil
				}
			}

			if n.Update != nil {
				_, err = i.evaluateNode(n.Update)
				if err != nil {
					return TypedValue{}, err
				}
			}

			condVal, err = i.evaluateNode(n.Condition)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *ConstantDeclaration:
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		if _, ok := i.getVar(n.Name); ok {
			return TypedValue{}, fmt.Errorf("variable '%s' already declared", n.Name)
		}
		i.setVar(n.Name, value)
		i.setConstant(n.Name)
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *HelpStatement:
		countryVal, err := i.evaluateNode(n.Country)
		if err != nil {
			return TypedValue{}, err
		}
		if countryVal.Type != TypeString {
			return TypedValue{}, fmt.Errorf("help argument must be a string")
		}
		country := countryVal.Value.(string)
		i.showHelp(country)
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *OrientationStatement:
		if isFileExecution {
			i.runOrientationDemo()
		} else {
			i.runOrientationTest()
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *FunctionDeclaration:
		i.mu.Lock()
		i.functions[n.Name] = n
		if n.Exported {
			i.exportedFuncs[n.Name] = n
		}
		i.mu.Unlock()
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *FunctionCall:
		return i.callFunction(n)
	case *ReturnStatement:
		value, err := i.evaluateNode(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		i.returnValue = value
		i.returnFlag = true
		return value, nil
	case *HeteroHomoStatement:
    var err error
    for _, stmt := range n.HeteroBlock {
        _, err = i.evaluateNode(stmt)
        if err != nil {
            break
        }
    }

    if err != nil {
        i.setVar("error", NewTypedString(err.Error()))
        i.setType("error", "lesbian")
        for _, stmt := range n.HomoBlock {
            _, catchErr := i.evaluateNode(stmt)
            if catchErr != nil {
                return TypedValue{}, catchErr
            }
        }
        return TypedValue{Type: TypeNull, Value: nil}, nil
    }
    return TypedValue{Type: TypeNull, Value: nil}, nil
	case *ExpressionStatement:
		val, err := i.evaluateNode(n.Expr)
		if err != nil {
			return TypedValue{}, err
		}
		return val, nil
	default:
		return TypedValue{Type: TypeNull, Value: nil}, nil
	}
}

func (i *Interpreter) handleInclude(node *IncludeStatement) (TypedValue, error) {
	filename := node.Filename
	if err := i.sandbox.CheckFilePath(filename); err != nil {
		return TypedValue{}, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return TypedValue{}, fmt.Errorf("cannot read include file '%s': %v", filename, err)
	}

	lexer := NewLexer(string(data))
	tokens, err := lexer.Tokenize()
	if err != nil {
		return TypedValue{}, fmt.Errorf("lexer error in include: %v", err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return TypedValue{}, fmt.Errorf("parser error in include: %v", err)
	}

	_, err = i.evaluateProgram(ast)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error in include: %v", err)
	}

	return TypedValue{Type: TypeNull, Value: nil}, nil
}

func (i *Interpreter) callFunction(n *FunctionCall) (TypedValue, error) {
	if fn, ok := i.getBuiltinFunction(n.Name); ok {
		args := make([]TypedValue, len(n.Args))
		for idx, arg := range n.Args {
			val, err := i.evaluateNode(arg)
			if err != nil {
				return TypedValue{}, err
			}
			args[idx] = val
		}
		return fn(args)
	}

	fn, ok := i.functions[n.Name]
	if !ok {
		return TypedValue{}, fmt.Errorf("function not defined: %s", n.Name)
	}

	if len(fn.Params) != len(n.Args) {
		return TypedValue{}, fmt.Errorf("argument count mismatch in call to %s: expected %d, got %d",
			n.Name, len(fn.Params), len(n.Args))
	}

	if i.recursionDepth > i.maxRecursion {
		return TypedValue{}, fmt.Errorf("maximum recursion depth exceeded: %d", i.maxRecursion)
	}
	i.recursionDepth++

	argValues := make([]TypedValue, len(n.Args))
	for idx, arg := range n.Args {
		val, err := i.evaluateNode(arg)
		if err != nil {
			i.recursionDepth--
			return TypedValue{}, err
		}
		argValues[idx] = val
	}

	i.pushFrame()

	for idx, param := range fn.Params {
		i.setVar(param, argValues[idx])
	}

	i.returnFlag = false
	i.returnValue = TypedValue{Type: TypeNull, Value: nil}

	var lastResult TypedValue
	var err error
	for _, stmt := range fn.Body {
		lastResult, err = i.evaluateNode(stmt)
		if err != nil {
			i.popFrame()
			i.recursionDepth--
			return TypedValue{}, err
		}
		if i.returnFlag {
			break
		}
	}

	result := i.returnValue
	if result.Type == TypeNull {
		result = lastResult
	}

	i.popFrame()
	i.recursionDepth--
	return result, nil
}

func (i *Interpreter) add(a, b TypedValue) (TypedValue, error) {
	if a.Type == TypeString || b.Type == TypeString {
		return NewTypedString(a.String() + b.String()), nil
	}

	if (a.Type == TypeInt || a.Type == TypeFloat) && (b.Type == TypeInt || b.Type == TypeFloat) {
		va, vb, err := i.toFloat(a, b)
		if err != nil {
			return TypedValue{}, err
		}
		if a.Type == TypeInt && b.Type == TypeInt {
			return NewTypedInt(int(va) + int(vb)), nil
		}
		return NewTypedFloat(va + vb), nil
	}

	return TypedValue{}, fmt.Errorf("cannot add %v and %v", a.Type, b.Type)
}

func (i *Interpreter) subtract(a, b TypedValue) (TypedValue, error) {
	if (a.Type != TypeInt && a.Type != TypeFloat) || (b.Type != TypeInt && b.Type != TypeFloat) {
		return TypedValue{}, fmt.Errorf("subtraction requires numeric operands")
	}
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return TypedValue{}, err
	}
	if a.Type == TypeInt && b.Type == TypeInt {
		return NewTypedInt(int(va) - int(vb)), nil
	}
	return NewTypedFloat(va - vb), nil
}

func (i *Interpreter) multiply(a, b TypedValue) (TypedValue, error) {
	if (a.Type != TypeInt && a.Type != TypeFloat) || (b.Type != TypeInt && b.Type != TypeFloat) {
		return TypedValue{}, fmt.Errorf("multiplication requires numeric operands")
	}
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return TypedValue{}, err
	}
	if a.Type == TypeInt && b.Type == TypeInt {
		return NewTypedInt(int(va) * int(vb)), nil
	}
	return NewTypedFloat(va * vb), nil
}

func (i *Interpreter) divide(a, b TypedValue) (TypedValue, error) {
	if (a.Type != TypeInt && a.Type != TypeFloat) || (b.Type != TypeInt && b.Type != TypeFloat) {
		return TypedValue{}, fmt.Errorf("division requires numeric operands")
	}
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return TypedValue{}, err
	}
	if vb == 0 {
		return TypedValue{}, fmt.Errorf("division by zero")
	}
	if a.Type == TypeInt && b.Type == TypeInt {
		return NewTypedInt(int(va) / int(vb)), nil
	}
	return NewTypedFloat(va / vb), nil
}

func (i *Interpreter) modulo(a, b TypedValue) (TypedValue, error) {
	if a.Type != TypeInt || b.Type != TypeInt {
		return TypedValue{}, fmt.Errorf("modulo requires integer operands")
	}
	va := a.Value.(int)
	vb := b.Value.(int)
	if vb == 0 {
		return TypedValue{}, fmt.Errorf("modulo by zero")
	}
	return NewTypedInt(va % vb), nil
}

func (i *Interpreter) toFloat(a, b TypedValue) (float64, float64, error) {
	var va, vb float64
	switch a.Type {
	case TypeInt:
		va = float64(a.Value.(int))
	case TypeFloat:
		va = a.Value.(float64)
	default:
		return 0, 0, fmt.Errorf("cannot convert to number: %v", a.Type)
	}
	switch b.Type {
	case TypeInt:
		vb = float64(b.Value.(int))
	case TypeFloat:
		vb = b.Value.(float64)
	default:
		return 0, 0, fmt.Errorf("cannot convert to number: %v", b.Type)
	}
	return va, vb, nil
}

func (i *Interpreter) isTruthy(val TypedValue) bool {
	switch val.Type {
	case TypeBool:
		return val.Value.(bool)
	case TypeInt:
		return val.Value.(int) != 0
	case TypeFloat:
		return val.Value.(float64) != 0
	case TypeString:
		return val.Value.(string) != ""
	case TypeArray:
		return len(val.Value.([]TypedValue)) > 0
	default:
		return false
	}
}

func (i *Interpreter) pushFrame() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.callStack = append(i.callStack, callFrame{
		vars:      make(map[string]TypedValue),
		types:     make(map[string]string),
		constants: make(map[string]bool),
	})
}

func (i *Interpreter) popFrame() {
	i.mu.Lock()
	defer i.mu.Unlock()
	if len(i.callStack) > 1 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
}

func (i *Interpreter) getVar(name string) (TypedValue, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	for idx := len(i.callStack) - 1; idx >= 0; idx-- {
		if val, ok := i.callStack[idx].vars[name]; ok {
			return val, true
		}
	}
	return TypedValue{Type: TypeNull, Value: nil}, false
}

func (i *Interpreter) setVar(name string, value TypedValue) {
	i.mu.Lock()
	defer i.mu.Unlock()
	top := len(i.callStack) - 1
	if i.callStack[top].constants[name] {
		panic(fmt.Errorf("cannot assign to constant '%s'", name))
	}
	i.callStack[top].vars[name] = value
}

func (i *Interpreter) getType(name string) (string, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	for idx := len(i.callStack) - 1; idx >= 0; idx-- {
		if typ, ok := i.callStack[idx].types[name]; ok {
			return typ, true
		}
	}
	return "", false
}

func (i *Interpreter) setType(name string, typ string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	top := len(i.callStack) - 1
	i.callStack[top].types[name] = typ
}

func (i *Interpreter) setConstant(name string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	top := len(i.callStack) - 1
	i.callStack[top].constants[name] = true
}

func (i *Interpreter) isConstant(name string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	for idx := len(i.callStack) - 1; idx >= 0; idx-- {
		if c, ok := i.callStack[idx].constants[name]; ok && c {
			return true
		}
	}
	return false
}

func (i *Interpreter) getTypeFromKeyword(keyword string) ValueType {
	switch keyword {
	case "lesbian":
		return TypeString
	case "gay":
		return TypeInt
	case "queer":
		return TypeFloat
	case "nonbinary":
		return TypeBool
	case "gender":
		return TypeArray
	default:
		return TypeNull
	}
}

func (i *Interpreter) checkType(typ string, value TypedValue) error {
	expectedType := i.getTypeFromKeyword(typ)
	if expectedType == TypeNull {
		return fmt.Errorf("unknown type: %s", typ)
	}
	if value.Type != expectedType {
		return fmt.Errorf("type mismatch: expected %s (%v), got %v", typ, expectedType, value.Type)
	}
	return nil
}

func (i *Interpreter) compareTypedValues(a, b TypedValue) (int, error) {
	if a.Type != b.Type {
		return 0, fmt.Errorf("cannot compare different types: %v and %v", a.Type, b.Type)
	}

	switch a.Type {
	case TypeInt:
		va := a.Value.(int)
		vb := b.Value.(int)
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	case TypeFloat:
		va := a.Value.(float64)
		vb := b.Value.(float64)
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	case TypeString:
		va := a.Value.(string)
		vb := b.Value.(string)
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot compare type %v", a.Type)
	}
}

func (i *Interpreter) evaluateBinaryOp(left TypedValue, op string, right TypedValue) (TypedValue, error) {
	switch op {
	case "+":
		return i.add(left, right)
	case "-":
		return i.subtract(left, right)
	case "*":
		return i.multiply(left, right)
	case "/":
		return i.divide(left, right)
	case "%":
		return i.modulo(left, right)
	case "==":
		if left.Type == right.Type {
			return NewTypedBool(reflect.DeepEqual(left.Value, right.Value)), nil
		}
		return NewTypedBool(false), nil
	case "!=":
		if left.Type == right.Type {
			return NewTypedBool(!reflect.DeepEqual(left.Value, right.Value)), nil
		}
		return NewTypedBool(true), nil
	case "<":
		cmp, err := i.compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp < 0), nil
	case ">":
		cmp, err := i.compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp > 0), nil
	case "<=":
		cmp, err := i.compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp <= 0), nil
	case ">=":
		cmp, err := i.compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp >= 0), nil
	case "&&":
		return NewTypedBool(i.isTruthy(left) && i.isTruthy(right)), nil
	case "||":
		return NewTypedBool(i.isTruthy(left) || i.isTruthy(right)), nil
	default:
		return TypedValue{}, fmt.Errorf("unknown operator: %s", op)
	}
}

// ============================================================
// НОВЫЕ ФУНКЦИИ - ПОЛНАЯ РЕАЛИЗАЦИЯ
// ============================================================

// 1. КАМИНГ-АУТ ПИСЬМО
func (i *Interpreter) comingOutLetter(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("comingOutLetter: expected 3 arguments (name, recipient, relationship)")
	}

	name, _ := args[0].Value.(string)
	recipient, _ := args[1].Value.(string)
	relationship, _ := args[2].Value.(string)

	style := "gentle"
	if len(args) > 3 {
		if s, ok := args[3].Value.(string); ok {
			style = s
		}
	}

	var letter string
	switch style {
	case "gentle":
		letter = fmt.Sprintf(`
Дорогой/ая %s,

Я хочу поделиться с тобой чем-то очень важным о себе. Я - ЛГБТ+ человек, и я открываюсь тебе, потому что ты значишь для меня многое.

Твое принятие и поддержка очень важны для меня. Я тот же человек, каким ты меня знал(а), просто теперь ты знаешь еще одну часть меня.

С любовью,
%s
`, recipient, name)
	case "direct":
		letter = fmt.Sprintf(`
%s,

Я - %s, и я ЛГБТ+. Я говорю тебе это напрямую, потому что уважаю тебя и хочу, чтобы между нами была честность.

Мне важна твоя поддержка, но даже если ты не готов(а) ее дать - я остаюсь собой.

С уважением,
%s
`, relationship, name, name)
	case "poetic":
		letter = fmt.Sprintf(`
Мой дорогой/ая %s,

В этом мире, полном красок,
Я нашел/нашла свой настоящий цвет.
И я хочу, чтобы ты знал(а),
Что этот цвет - часть меня.

Я - ЛГБТ+ человек,
И я открываю тебе свое сердце,
Потому что верю в нашу связь.

С нежностью,
%s
`, recipient, name)
	default:
		letter = fmt.Sprintf(`
%s,

Я хочу сказать тебе, что я %s, и я ЛГБТ+.
Ты важен/важна для меня, поэтому я делюсь с тобой этим.

С уважением,
%s
`, recipient, name, name)
	}

	return NewTypedString(letter), nil
}

// 2. ТРАНС-ПОДДЕРЖКА
func (i *Interpreter) transSupport(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("transSupport: expected query argument")
	}

	query, _ := args[0].Value.(string)

	supportMap := map[string]string{
		"гормон":   "Информация о гормональной терапии:\n• Обратитесь к эндокринологу\n• Получите направление от психолога\n• Список транс-дружественных врачей доступен по запросу",
		"документ": "Информация о смене документов:\n• Заявление в ЗАГС\n• Медицинское заключение\n• Решение суда\n• Паспорт с новым полом",
		"психолог": "Транс-дружественные психологи:\n• Поиск по вашему городу\n• Бесплатные консультации для транс-людей\n• Группы поддержки",
		"право":    "Права транс-людей:\n• Защита от дискриминации\n• Право на медицинскую помощь\n• Право на смену документов",
	}

	result := "🏳️‍⚧️ Транс-поддержка:\n"
	if info, ok := supportMap[strings.ToLower(query)]; ok {
		result += info + "\n"
	} else {
		result += fmt.Sprintf("ℹ️ Информация по запросу '%s' уточняется.\n", query)
	}
	result += "\n💙 Вы не одиноки! Поддержка доступна 24/7."
	return NewTypedString(result), nil
}

// 3. ПОИСК СОЮЗНИКОВ
func (i *Interpreter) findAllies(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("findAllies: expected place and radius")
	}

	place, _ := args[0].Value.(string)
	radius, _ := args[1].Value.(int)

	allies := []string{
		"🏳️‍🌈 Центр 'Радуга' - ул. Мира, 15 (всегда открыты для поддержки)",
		"🏳️‍🌈 Кафе 'Вместе' - пр. Свободы, 42 (безопасное пространство)",
		"🏳️‍🌈 Книжный 'Открытый мир' - ул. Ленина, 7 (ЛГБТ+ литература)",
		"🏳️‍🌈 Арт-пространство 'Толерантность' - пер. Художников, 5",
		"🏳️‍🌈 Спортклуб 'Единство' - ул. Спортивная, 23",
	}

	result := fmt.Sprintf("📍 Союзники в %s (радиус %d км):\n", place, radius)
	for _, a := range allies {
		result += "  • " + a + "\n"
	}
	result += "\n🌟 Помните: вы не одиноки! Союзники всегда рядом."
	return NewTypedString(result), nil
}

// 4. ИСТОРИЯ ЛГБТ+
func (i *Interpreter) lgbtHistory(args []TypedValue) (TypedValue, error) {
	year := 1969
	if len(args) > 0 {
		if y, ok := args[0].Value.(int); ok {
			year = y
		}
	}

	country := ""
	if len(args) > 1 {
		if c, ok := args[1].Value.(string); ok {
			country = c
		}
	}

	historyEvents := map[int]string{
		1924:  "Основана первая ЛГБТ-организация в США (Society for Human Rights)",
		1950:  "Создана первая гей-организация в Германии",
		1969:  "Стоунволлские бунты - начало движения за права ЛГБТ+",
		1973:  "Американская психиатрическая ассоциация исключила гомосексуальность из списка психических расстройств",
		1982:  "Первая публикация о ВИЧ/СПИДе в ЛГБТ-сообществе",
		1990:  "ВОЗ исключила гомосексуальность из списка болезней",
		1994:  "Основан первый ЛГБТ-журнал в России",
		2001:  "Нидерланды - первая страна, легализовавшая однополые браки",
		2003:  "Легализация однополых браков в Бельгии и Канаде",
		2005:  "Испания легализует однополые браки",
		2006:  "Южная Африка - первая африканская страна, легализовавшая однополые браки",
		2010:  "Португалия, Исландия и Аргентина легализуют однополые браки",
		2013:  "Франция и Новая Зеландия легализуют однополые браки",
		2015:  "Верховный суд США легализует однополые браки во всех штатах",
		2017:  "Германия и Мальта легализуют однополые браки",
		2019:  "Тайвань - первая азиатская страна, легализовавшая однополые браки",
		2020:  "В России принят закон о запрете пропаганды ЛГБТ+",
		2022:  "Чили легализует однополые браки",
		2023:  "В Греции легализованы однополые браки и усыновление",
		2024:  "Болгария и Румыния принимают законы о гражданских партнерствах",
	}

	result := fmt.Sprintf("📅 ЛГБТ+ история (%d год", year)
	if country != "" {
		result += fmt.Sprintf(", %s", country)
	}
	result += "):\n"

	found := false
	for y, event := range historyEvents {
		if (country == "" || strings.Contains(strings.ToLower(event), strings.ToLower(country))) &&
			(year == 0 || y == year) {
			result += fmt.Sprintf("  • %d: %s\n", y, event)
			found = true
		}
	}

	if !found {
		if year > 0 {
			result += fmt.Sprintf("  ℹ️ Нет событий в %d году\n", year)
		} else {
			result += "  ℹ️ Нет событий по вашему запросу\n"
		}
	}

	result += "\n🏳️‍🌈 Помните: мы сильны, потому что мы помним свою историю!"
	return NewTypedString(result), nil
}

// 5. АВАТАР ПОДДЕРЖКИ
func (i *Interpreter) prideAvatar(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("prideAvatar: expected name and style")
	}

	name, _ := args[0].Value.(string)
	style, _ := args[1].Value.(string)

	flagTypes := []string{"pride", "trans", "lesbian", "gay", "bisexual", "nonbinary"}
	selectedFlag := flagTypes[i.rand.Intn(len(flagTypes))]

	if style != "" {
		for _, f := range flagTypes {
			if strings.Contains(strings.ToLower(style), f) {
				selectedFlag = f
				break
			}
		}
	}

	colors := i.flagColors[selectedFlag]
	if len(colors) == 0 {
		colors = i.flagColors["pride"]
	}

	avatar := fmt.Sprintf(`
🏳️‍🌈 Аватар для %s (стиль: %s, флаг: %s)
🌈 Цвета: %s

  /\\_/\\
 ( o.o )  💙 Ты важен(на)!
  > ^ <

✨ Твоя идентичность - твоя сила!
🏳️‍🌈 Гордись собой!
`, name, style, selectedFlag, strings.Join(colors, ", "))

	return NewTypedString(avatar), nil
}

// 6. ДНЕВНИК ЛГБТ+
func (i *Interpreter) lgbtJournal(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("lgbtJournal: expected action and text")
	}

	action, _ := args[0].Value.(string)
	text, _ := args[1].Value.(string)

	switch strings.ToLower(action) {
	case "add":
		encrypted, err := i.encryption.Encrypt(text)
		if err != nil {
			return TypedValue{}, err
		}
		mood := "neutral"
		if len(args) > 2 {
			if m, ok := args[2].Value.(string); ok {
				mood = m
			}
		}
		entry := JournalEntry{
			Timestamp: time.Now(),
			Text:      encrypted,
			Mood:      mood,
			Encrypted: true,
		}
		i.journalMu.Lock()
		i.journal = append(i.journal, entry)
		i.journalMu.Unlock()
		return NewTypedString("✅ Запись добавлена в дневник (зашифрована)"), nil

	case "read":
		decrypted, err := i.encryption.Decrypt(text)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedString(fmt.Sprintf("📖 Расшифрованная запись:\n%s", decrypted)), nil

	case "list":
		i.journalMu.RLock()
		defer i.journalMu.RUnlock()

		result := fmt.Sprintf("📖 Записей в дневнике: %d\n", len(i.journal))
		for idx, entry := range i.journal {
			decrypted, err := i.encryption.Decrypt(entry.Text)
			if err != nil {
				decrypted = "[Ошибка расшифровки]"
			}
			if len(decrypted) > 50 {
				decrypted = decrypted[:50] + "..."
			}
			result += fmt.Sprintf("  %d. [%s] %s (настроение: %s)\n",
				idx+1, entry.Timestamp.Format("2006-01-02 15:04"), decrypted, entry.Mood)
		}
		return NewTypedString(result), nil

	default:
		return TypedValue{}, fmt.Errorf("lgbtJournal: unknown action '%s'", action)
	}
}

// 7. КАЛЕНДАРЬ СОБЫТИЙ
func (i *Interpreter) getLGBTCEvents(args []TypedValue) (TypedValue, error) {
	country := "all"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	
	events := []EventEntry{
		{ID: "e1", Title: "Международный день ЛГБТ+", Date: "2024-05-17", Country: "all", City: "Международный", Type: "activism", Description: "День борьбы с гомофобией", URL: "https://may17.org", Organizer: "UNESCO", Tags: []string{"права", "активизм"}},
		{ID: "e2", Title: "Московский прайд", Date: "2024-06-01", Country: "Россия", City: "Москва", Type: "pride", Description: "Ежегодный марш гордости", URL: "https://pride.ru", Organizer: "ЛГБТ-сеть", Tags: []string{"прайд", "активизм"}},
		{ID: "e3", Title: "Петербургский прайд", Date: "2024-06-15", Country: "Россия", City: "Санкт-Петербург", Type: "pride", Description: "Марш гордости в СПб", URL: "https://spbpride.ru", Organizer: "Сфера", Tags: []string{"прайд", "активизм"}},
		{ID: "e4", Title: "World Pride 2024", Date: "2024-07-01", Country: "Германия", City: "Берлин", Type: "pride", Description: "Всемирный прайд", URL: "https://worldpride.org", Organizer: "International", Tags: []string{"прайд", "международный"}},
		{ID: "e5", Title: "Транс-марш", Date: "2024-08-15", Country: "США", City: "Нью-Йорк", Type: "activism", Description: "Марш за права транс-людей", URL: "https://transmarch.org", Organizer: "Trans Coalition", Tags: []string{"транс", "активизм"}},
		{ID: "e6", Title: "ЛГБТ+ конференция", Date: "2024-09-20", Country: "Великобритания", City: "Лондон", Type: "education", Description: "Международная конференция", URL: "https://lgbtconf.org", Organizer: "ILGA", Tags: []string{"конференция", "образование"}},
		{ID: "e7", Title: "Российская ЛГБТ+ неделя", Date: "2024-10-01", Country: "Россия", City: "Москва", Type: "support", Description: "Неделя поддержки", URL: "https://lgbtweek.ru", Organizer: "ЛГБТ-сеть", Tags: []string{"поддержка", "образование"}},
		{ID: "e8", Title: "Международный день транс-людей", Date: "2024-11-20", Country: "all", City: "Международный", Type: "support", Description: "День памяти транс-людей", URL: "https://tdor.org", Organizer: "International", Tags: []string{"транс", "память"}},
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📅 ЛГБТ+ события (страна: %s):\n", country))

	for _, event := range events {
		if country != "all" && !strings.Contains(strings.ToLower(event.Country), strings.ToLower(country)) {
			continue
		}
		result.WriteString(fmt.Sprintf("  • %s - %s (%s, %s)\n", event.Date, event.Title, event.City, event.Country))
		result.WriteString(fmt.Sprintf("    📝 %s\n", event.Description))
		result.WriteString(fmt.Sprintf("    🔗 %s\n", event.URL))
	}

	return NewTypedString(result.String()), nil
}

// 8. СОЗДАНИЕ ФЛАГА
func (i *Interpreter) createFlag(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("createFlag: expected flag type")
	}

	flagType, _ := args[0].Value.(string)
	size := 10
	if len(args) > 1 {
		if s, ok := args[1].Value.(int); ok {
			size = s
		}
	}

	flagType = strings.ToLower(flagType)
	colors, ok := i.flagColors[flagType]
	if !ok {
		colors = i.flagColors["pride"]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🏳️‍🌈 Флаг '%s' (размер %d):\n", flagType, size))
	for _, color := range colors {
		for j := 0; j < size; j++ {
			result.WriteString("━")
		}
		result.WriteString(fmt.Sprintf(" %s\n", color))
	}
	result.WriteString(fmt.Sprintf("\n💡 Значение флага '%s':\n", flagType))
	result.WriteString(getFlagDescription(flagType))

	return NewTypedString(result.String()), nil
}

func getFlagDescription(flagType string) string {
	descriptions := map[string]string{
		"pride":     "Радужный флаг - символ гордости ЛГБТ+ сообщества, обозначает разнообразие и инклюзивность.",
		"trans":     "Транс-флаг - символ транс-сообщества: голубой - мальчики, розовый - девочки, белый - небинарные люди.",
		"lesbian":   "Лесбийский флаг - символ лесбийского сообщества.",
		"gay":       "Гей-флаг - символ гей-сообщества.",
		"bisexual":  "Бисексуальный флаг - розовый (однополые), синий (противоположные), фиолетовый (все).",
		"nonbinary": "Небинарный флаг - желтый (небинарность), белый (множество гендеров), фиолетовый (смешение), черный (отсутствие гендера).",
	}
	if desc, ok := descriptions[flagType]; ok {
		return desc
	}
	return "Флаг ЛГБТ+ сообщества."
}

// 9. ТЕРАПЕВТИЧЕСКИЕ УПРАЖНЕНИЯ
func (i *Interpreter) lgbtTherapy(args []TypedValue) (TypedValue, error) {
	exerciseType := "acceptance"
	if len(args) > 0 {
		if t, ok := args[0].Value.(string); ok {
			exerciseType = t
		}
	}

	duration := 5
	if len(args) > 1 {
		if d, ok := args[1].Value.(int); ok {
			duration = d
		}
	}

	exercises := map[string]string{
		"acceptance": "Упражнение на самопринятие:\n1. Сядьте удобно\n2. Закройте глаза\n3. Повторяйте: 'Я принимаю себя таким(ой), какой(ая) я есть'\n4. Чувствуйте тепло в сердце\n5. Дышите глубоко",
		"anxiety":    "Упражнение при тревоге:\n1. Дышите медленно (4-4-6)\n2. Назовите 5 вещей, которые видите\n3. Назовите 4 вещи, которые слышите\n4. Назовите 3 вещи, которые чувствуете\n5. Назовите 2 вещи, которые вдыхаете\n6. Назовите 1 вещь, которую пробуете",
		"strength":   "Упражнение на силу:\n1. Вспомните, что вы пережили\n2. Почувствуйте свою силу\n3. Скажите: 'Я сильнее, чем думаю'\n4. Улыбнитесь себе в зеркало",
		"gratitude":  "Упражнение на благодарность:\n1. Напишите 3 вещи, за которые вы благодарны\n2. Напишите 3 вещи, которые вам нравятся в себе\n3. Напишите 3 человека, которые вас поддерживают",
	}

	result := fmt.Sprintf("🧘 Терапевтическое упражнение (%s, %d минут):\n", exerciseType, duration)
	if exercise, ok := exercises[exerciseType]; ok {
		result += exercise + "\n"
	} else {
		result += "ℹ️ Упражнение не найдено. Доступны: acceptance, anxiety, strength, gratitude\n"
	}
	result += "\n💙 Берегите себя! Вы важны."

	return NewTypedString(result), nil
}

// 10. ЦИТАТЫ ДНЯ
func (i *Interpreter) lgbtQuote(args []TypedValue) (TypedValue, error) {
	theme := ""
	if len(args) > 0 {
		if t, ok := args[0].Value.(string); ok {
			theme = t
		}
	}

	idx := i.rand.Intn(len(i.quotes))
	quote := i.quotes[idx]

	if theme != "" {
		for _, q := range i.quotes {
			if strings.Contains(strings.ToLower(q), strings.ToLower(theme)) {
				return NewTypedString(fmt.Sprintf("💬 Цитата дня (тема: %s):\n%s", theme, q)), nil
			}
		}
	}

	return NewTypedString(fmt.Sprintf("💬 Цитата дня:\n%s", quote)), nil
}

// 11. КНИЖНЫЙ КЛУБ
func (i *Interpreter) lgbtBookClub(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		result := "📚 ЛГБТ+ литература:\n"
		for _, book := range i.books {
			result += fmt.Sprintf("  • '%s' - %s (%s, %s)\n", book.Title, book.Author, book.Year, book.Genre)
			result += fmt.Sprintf("    📝 %s\n", book.Description)
			result += fmt.Sprintf("    🏷️ Темы: %s\n", strings.Join(book.Themes, ", "))
		}
		return NewTypedString(result), nil
	}

	title, _ := args[0].Value.(string)
	for _, book := range i.books {
		if strings.Contains(strings.ToLower(book.Title), strings.ToLower(title)) {
			result := fmt.Sprintf("📖 %s\n", book.Title)
			result += fmt.Sprintf("👤 Автор: %s\n", book.Author)
			result += fmt.Sprintf("📅 Год: %s\n", book.Year)
			result += fmt.Sprintf("📚 Жанр: %s\n", book.Genre)
			result += fmt.Sprintf("📝 Описание: %s\n", book.Description)
			result += fmt.Sprintf("🏷️ Темы: %s\n", strings.Join(book.Themes, ", "))
			result += fmt.Sprintf("👥 Возраст: %s\n", book.AgeGroup)
			return NewTypedString(result), nil
		}
	}

	return NewTypedString(fmt.Sprintf("ℹ️ Книга '%s' не найдена", title)), nil
}

// 12. ОПРЕДЕЛИТЕЛЬ ПРИВИЛЕГИЙ
func (i *Interpreter) checkPrivilege(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("checkPrivilege: expected country")
	}

	country, _ := args[0].Value.(string)

	privilegeMap := map[string]int{
		"Россия":         3,
		"США":            8,
		"Германия":       9,
		"Франция":        8,
		"Великобритания": 8,
		"Канада":         9,
		"Нидерланды":     9,
		"Испания":        8,
		"Португалия":     7,
		"Аргентина":      6,
		"ЮАР":            5,
		"Израиль":        7,
		"Турция":         2,
		"Китай":          2,
		"Япония":         5,
		"Корея":          4,
		"Бразилия":       6,
		"Мексика":        4,
	}

	score, ok := privilegeMap[country]
	if !ok {
		score = 5
	}

	level := "низкий"
	if score >= 8 {
		level = "высокий"
	} else if score >= 6 {
		level = "средний"
	} else if score >= 4 {
		level = "ниже среднего"
	}

	emoji := "⚠️"
	if score >= 8 {
		emoji = "🌟"
	} else if score >= 6 {
		emoji = "👍"
	} else if score >= 4 {
		emoji = "⚖️"
	}

	result := fmt.Sprintf("%s Уровень безопасности для ЛГБТ+ в %s:\n", emoji, country)
	result += fmt.Sprintf("📊 Рейтинг: %d/10 (%s уровень)\n", score, level)
	result += fmt.Sprintf("📝 Рекомендация: %s\n", getSafetyRecommendation(score))

	return NewTypedString(result), nil
}

func getSafetyRecommendation(score int) string {
	if score >= 8 {
		return "Относительно безопасно. Сообщество поддерживает ЛГБТ+ права."
	} else if score >= 6 {
		return "В целом безопасно, но есть некоторые ограничения."
	} else if score >= 4 {
		return "Будьте осторожны в некоторых районах. Ищите союзников."
	}
	return "Высокий риск. Рекомендуется быть крайне осторожным."
}

// 13. СЕРТИФИКАТ БЕЗОПАСНОСТИ
func (i *Interpreter) createSafeSpaceCert(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("createSafeSpaceCert: expected school and address")
	}

	school, _ := args[0].Value.(string)
	address, _ := args[1].Value.(string)

	cert := fmt.Sprintf(`
🏳️‍🌈 СЕРТИФИКАТ "БЕЗОПАСНОЕ ПРОСТРАНСТВО" 🏳️‍🌈

Настоящим подтверждается, что

    %s
    %s

является безопасным пространством для ЛГБТ+ студентов и сотрудников.

Это место гарантирует:
• Принятие всех гендерных идентичностей
• Нулевую толерантность к дискриминации
• Конфиденциальность и поддержку
• Доступ к ресурсам ЛГБТ+

Дата: %s
Сертификат действителен до: %s

💙 Вместе мы создаем безопасный мир!
`, school, address, time.Now().Format("2006-01-02"), time.Now().AddDate(1, 0, 0).Format("2006-01-02"))

	return NewTypedString(cert), nil
}

// 14. ПЛАН АКТИВИЗМА
func (i *Interpreter) activismPlan(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("activismPlan: expected goal and country")
	}

	goal, _ := args[0].Value.(string)
	country, _ := args[1].Value.(string)

	plan := fmt.Sprintf(`
✊ АКТИВИЗМ-ПЛАН

Цель: %s
Страна: %s

Шаги:
1. 📚 Изучите законодательство о ЛГБТ+ в стране
2. 🤝 Найдите местные ЛГБТ+ организации
3. 📝 Создайте петицию или письмо
4. 🌐 Используйте социальные сети для информирования
5. 🗣️ Организуйте встречи и дискуссии
6. 📊 Отслеживайте прогресс и достижения

Важно:
• Соблюдайте местные законы
• Будьте в безопасности
• Поддерживайте других активистов

💪 Вместе мы можем изменить мир!
`, goal, country)

	return NewTypedString(plan), nil
}

// 15. МАНДАЛА ПОДДЕРЖКИ
func (i *Interpreter) prideMandalas(args []TypedValue) (TypedValue, error) {
	name := "друг"
	if len(args) > 0 {
		if n, ok := args[0].Value.(string); ok {
			name = n
		}
	}

	colors := []string{"❤️", "🧡", "💛", "💚", "💙", "💜"}
	if len(args) > 1 {
		if c, ok := args[1].Value.([]TypedValue); ok {
			colors = make([]string, len(c))
			for i, v := range c {
				if s, ok := v.Value.(string); ok {
					colors[i] = s
				}
			}
		}
	}

	mandala := fmt.Sprintf(`
🌸 МАНДАЛА ДЛЯ %s 🌸

    %s %s %s
   %s %s %s %s %s
  %s %s %s %s %s %s %s
   %s %s %s %s %s
    %s %s %s

💫 Каждый цвет — это часть твоей души.
🌈 Ты — прекрасное создание!
`, strings.ToUpper(name),
		colors[0], colors[1], colors[2],
		colors[3], colors[0], colors[4], colors[1], colors[5],
		colors[2], colors[3], colors[5], colors[0], colors[4], colors[1],
		colors[5], colors[2], colors[3], colors[0], colors[1],
		colors[4], colors[5], colors[2])

	return NewTypedString(mandala), nil
}

// 16. ХРОНОГРАФ БОРЬБЫ
func (i *Interpreter) lgbtTimeline(args []TypedValue) (TypedValue, error) {
	var result strings.Builder
	result.WriteString("🏳️‍🌈 Хронограф борьбы за права ЛГБТ+:\n\n")

	timeline := []string{
		"1924 - Основана первая ЛГБТ-организация",
		"1950 - Создана первая гей-организация в Германии",
		"1969 - Стоунволлские бунты",
		"1973 - Гомосексуальность исключена из списка психических расстройств в США",
		"1990 - ВОЗ исключает гомосексуальность из списка болезней",
		"1994 - Основан первый ЛГБТ-журнал в России",
		"2001 - Нидерланды - первая страна с однополыми браками",
		"2003 - Бельгия и Канада легализуют однополые браки",
		"2005 - Испания легализует однополые браки",
		"2006 - Южная Африка - первая африканская страна",
		"2010 - Португалия, Исландия, Аргентина",
		"2013 - Франция и Новая Зеландия",
		"2015 - Верховный суд США легализует однополые браки",
		"2017 - Германия и Мальта",
		"2019 - Тайвань - первая азиатская страна",
		"2022 - Чили легализует однополые браки",
		"2023 - Греция легализует однополые браки",
	}

	for _, event := range timeline {
		result.WriteString(fmt.Sprintf("  • %s\n", event))
	}

	result.WriteString("\n✊ Мы боремся, мы побеждаем! Наша история - наша сила!")
	return NewTypedString(result.String()), nil
}

// 17. КАРТА ХЕЙТА
func (i *Interpreter) hateMap(args []TypedValue) (TypedValue, error) {
	region := "все"
	if len(args) > 0 {
		if r, ok := args[0].Value.(string); ok {
			region = r
		}
	}

	hateLevels := map[string]string{
		"Россия":          "⚠️ Высокий уровень. Будьте осторожны.",
		"США":             "📊 Средний уровень. Региональные различия.",
		"Германия":        "✅ Низкий уровень. Относительно безопасно.",
		"Франция":         "✅ Низкий уровень. Безопасно.",
		"Великобритания":  "✅ Низкий уровень. Безопасно.",
		"Канада":          "✅ Очень низкий уровень. Очень безопасно.",
		"Турция":          "⚠️ Высокий уровень. Будьте осторожны.",
		"Китай":           "⚠️ Высокий уровень. Соблюдайте осторожность.",
		"Япония":          "📊 Средний уровень. Относительно безопасно.",
		"Бразилия":        "📊 Средний уровень. Осторожно в некоторых районах.",
		"ЮАР":             "📊 Средний уровень. Будьте внимательны.",
		"Австралия":       "✅ Низкий уровень. Безопасно.",
		"Новая Зеландия":  "✅ Низкий уровень. Очень безопасно.",
	}

	result := fmt.Sprintf("🗺️ Карта ненависти (%s):\n", region)
	if region == "все" {
		for country, level := range hateLevels {
			result += fmt.Sprintf("  • %s: %s\n", country, level)
		}
	} else {
		if level, ok := hateLevels[region]; ok {
			result += fmt.Sprintf("  • %s: %s\n", region, level)
		} else {
			result += fmt.Sprintf("  ℹ️ Информация по региону '%s' уточняется\n", region)
		}
	}

	result += "\n🛡️ Помните: вы не одиноки! Поддержка всегда рядом."
	return NewTypedString(result), nil
}

// 18. ПОИСК МЕНТОРА
func (i *Interpreter) findMentor(args []TypedValue) (TypedValue, error) {
	interest := "comingout"
	if len(args) > 0 {
		if in, ok := args[0].Value.(string); ok {
			interest = in
		}
	}

	experience := "beginner"
	if len(args) > 1 {
		if exp, ok := args[1].Value.(string); ok {
			experience = exp
		}
	}

	mentors := map[string]string{
		"comingout":  "👤 Ментор по каминг-ауту:\nАлекс (35 лет, гей)\n'Я помогу тебе найти слова и смелость'\n🕒 3 года опыта",
		"trans":      "🏳️‍⚧️ Ментор по транс-вопросам:\nДжей (28 лет, транс-женщина)\n'Я прошла этот путь и готова помочь'\n🕒 5 лет опыта",
		"anxiety":    "🧠 Ментор по тревоге:\nСаша (32 года, небинарный(ая))\n'Я научу тебя справляться'\n🕒 4 года опыта",
		"career":     "💼 Ментор по карьере:\nМария (40 лет, лесбиянка)\n'Я помогу тебе построить карьеру'\n🕒 10 лет опыта",
		"relationships": "❤️ Ментор по отношениям:\nДмитрий (30 лет, бисексуал)\n'Я помогу тебе найти любовь'\n🕒 3 года опыта",
	}

	result := fmt.Sprintf("🤝 Ментор по теме '%s' (уровень: %s):\n", interest, experience)
	if mentor, ok := mentors[interest]; ok {
		result += mentor + "\n"
	} else {
		result += "ℹ️ Ментор не найден. Попробуйте: comingout, trans, anxiety, career, relationships\n"
	}
	result += "\n🌟 Помните: у каждого есть свой путь, и вы не одни!"

	return NewTypedString(result), nil
}

// ============================================================
// ВСТРОЕННЫЕ ФУНКЦИИ - РЕГИСТРАЦИЯ ВСЕХ ФУНКЦИЙ
// ============================================================

func (i *Interpreter) getBuiltinFunction(name string) (func([]TypedValue) (TypedValue, error), bool) {
	switch name {
	// Оригинальные функции
	case "msg":
		return i.msg, true
	case "antiHomoPhobe":
		return i.antiHomoPhobe, true
	case "lgbtImg":
		return i.lgbtImg, true
	case "getLGBTResources":
		return i.getLGBTResources, true
	case "findSafeSpace":
		return i.findSafeSpace, true
	case "getCrisisSupport":
		return i.getCrisisSupport, true
	case "getLGBTQLaws":
		return i.getLGBTQLaws, true
	case "getDailyAffirmation":
		return i.getDailyAffirmation, true
	case "moodCheck":
		return i.moodCheck, true
	case "guidedBreathing":
		return i.guidedBreathing, true
	case "defineTerm":
		return i.defineTerm, true

	// Функции фильтра ненависти
	case "checkHate":
		return i.checkHate, true
	case "filterHate":
		return i.filterHate, true
	case "enableHateFilter":
		return i.enableHateFilter, true
	case "disableHateFilter":
		return i.disableHateFilter, true

	// НОВЫЕ ФУНКЦИИ (18+)
	case "comingOutLetter":
		return i.comingOutLetter, true
	case "transSupport":
		return i.transSupport, true
	case "findAllies":
		return i.findAllies, true
	case "lgbtHistory":
		return i.lgbtHistory, true
	case "prideAvatar":
		return i.prideAvatar, true
	case "lgbtJournal":
		return i.lgbtJournal, true
	case "getLGBTCEvents":
		return i.getLGBTCEvents, true
	case "createFlag":
		return i.createFlag, true
	case "lgbtTherapy":
		return i.lgbtTherapy, true
	case "lgbtQuote":
		return i.lgbtQuote, true
	case "lgbtBookClub":
		return i.lgbtBookClub, true
	case "checkPrivilege":
		return i.checkPrivilege, true
	case "createSafeSpaceCert":
		return i.createSafeSpaceCert, true
	case "activismPlan":
		return i.activismPlan, true
	case "prideMandalas":
		return i.prideMandalas, true
	case "lgbtTimeline":
		return i.lgbtTimeline, true
	case "hateMap":
		return i.hateMap, true
	case "findMentor":
		return i.findMentor, true

	default:
		return nil, false
	}
}

// ============================================================
// ОСТАЛЬНЫЕ ФУНКЦИИ (msg, antiHomoPhobe, lgbtImg, и т.д.)
// ============================================================

func (i *Interpreter) msg(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("msg: expected 2 arguments (title, text)")
	}

	title, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("msg: first argument (title) must be string")
	}

	text, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("msg: second argument (text) must be string")
	}

	hasHateTitle, filteredTitle := i.checkHateSpeech(title)
	if hasHateTitle {
		title = filteredTitle
		fmt.Fprintln(output, "⚠️ Обнаружена ненависть в заголовке диалога! Заголовок отфильтрован.")
	}

	hasHateText, filteredText := i.checkHateSpeech(text)
	if hasHateText {
		text = filteredText
		fmt.Fprintln(output, "⚠️ Обнаружена ненависть в тексте диалога! Текст отфильтрован.")
	}

	flags := MB_OK
	if len(args) > 2 {
		if typeStr, ok := args[2].Value.(string); ok {
			switch strings.ToLower(typeStr) {
			case "ok":
				flags = MB_OK
			case "okcancel":
				flags = MB_OKCANCEL
			case "yesno":
				flags = MB_YESNO
			case "yesnocancel":
				flags = MB_YESNOCANCEL
			case "retrycancel":
				flags = MB_RETRYCANCEL
			case "abortretryignore":
				flags = MB_ABORTRETRYIGNORE
			default:
				flags = MB_OK
			}
		}
	}

	if len(args) > 3 {
		if iconStr, ok := args[3].Value.(string); ok {
			switch strings.ToLower(iconStr) {
			case "info", "information":
				flags |= MB_ICONINFORMATION
			case "warning":
				flags |= MB_ICONWARNING
			case "error":
				flags |= MB_ICONERROR
			case "question":
				flags |= MB_ICONQUESTION
			}
		}
	}

	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return TypedValue{}, fmt.Errorf("msg: failed to create title: %v", err)
	}

	textPtr, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return TypedValue{}, fmt.Errorf("msg: failed to create text: %v", err)
	}

	result := MessageBox(0, textPtr, titlePtr, uint32(flags))

	var resultStr string
	switch result {
	case IDOK:
		resultStr = "ok"
	case IDCANCEL:
		resultStr = "cancel"
	case IDYES:
		resultStr = "yes"
	case IDNO:
		resultStr = "no"
	case IDRETRY:
		resultStr = "retry"
	case IDABORT:
		resultStr = "abort"
	case IDIGNORE:
		resultStr = "ignore"
	default:
		resultStr = "unknown"
	}

	response := "✅ Диалоговое окно показано (нажато: " + resultStr + ")"
	if hasHateTitle || hasHateText {
		response += " [фильтрация ненависти применена]"
	}
	return NewTypedString(response), nil
}

func (i *Interpreter) antiHomoPhobe(args []TypedValue) (TypedValue, error) {
	duration := 7
	if len(args) > 0 {
		if d, ok := args[0].Value.(int); ok {
			if d > 0 && d <= 30 {
				duration = d
			}
		}
	}

	fmt.Fprintf(output, "\n🚨🚨🚨 ANTI-GOMOPHOBE АКТИВИРОВАНА! 🚨🚨🚨\n")
	fmt.Fprintf(output, "🔄 Выдвигаю дисковод и мигаю лампочками клавиатуры...\n")
	fmt.Fprintf(output, "⏱️  Длительность: %d секунд\n\n", duration)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command", "(New-Object -ComObject Shell.Application).EjectCD()")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(output, "⚠️ Не удалось выдвинуть дисковод: %v\n", err)
		} else {
			fmt.Fprintf(output, "💿 Дисковод выдвинут\n")
		}
	} else {
		cmd := exec.Command("eject")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(output, "⚠️ Не удалось выдвинуть дисковод: %v\n", err)
		} else {
			fmt.Fprintf(output, "💿 Дисковод выдвинут\n")
		}
	}

	startTime := time.Now()
	done := make(chan bool)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		state := false
		for {
			select {
			case <-ticker.C:
				state = !state
				elapsed := time.Since(startTime).Seconds()
				remaining := duration - int(elapsed)
				if remaining < 0 {
					remaining = 0
				}

				if state {
					fmt.Fprintf(output, "\r💡💡💡 [Caps Lock] [Num Lock] [Scroll Lock] - Осталось: %2d сек   ", remaining)
				} else {
					fmt.Fprintf(output, "\r⬜⬜⬜ [Caps Lock] [Num Lock] [Scroll Lock] - Осталось: %2d сек   ", remaining)
				}

				if runtime.GOOS == "windows" {
					keybdEvent(0x14, 0, 0, 0)
					keybdEvent(0x14, 0, 0x0002, 0)
					keybdEvent(0x90, 0, 0, 0)
					keybdEvent(0x90, 0, 0x0002, 0)
					keybdEvent(0x91, 0, 0, 0)
					keybdEvent(0x91, 0, 0x0002, 0)
				}

				if elapsed >= float64(duration) {
					done <- true
					return
				}
			case <-done:
				return
			}
		}
	}()

	<-done
	fmt.Fprintln(output)

	if runtime.GOOS == "windows" {
		fmt.Fprintf(output, "📀 Для закрытия дисковода нажмите кнопку на приводе или используйте 'Извлечь' в проводнике.\n")
	} else {
		cmd := exec.Command("eject", "-t")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(output, "⚠️ Не удалось закрыть дисковод: %v\n", err)
		} else {
			fmt.Fprintf(output, "📀 Дисковод закрыт\n")
		}
	}

	fmt.Fprintf(output, "\n✅ Anti-GomoPhobe завершена!\n")
	fmt.Fprintf(output, "📊 Лампочки мигали %d секунд\n", duration)

	result := fmt.Sprintf("✅ Anti-GomoPhobe выполнена! Дисковод выдвинут на %d секунд, лампочки мигали.", duration)
	return NewTypedString(result), nil
}

func keybdEvent(bVk byte, bScan byte, dwFlags uintptr, dwExtraInfo uintptr) {
	procKeybdEvent.Call(uintptr(bVk), uintptr(bScan), dwFlags, dwExtraInfo)
}

func (i *Interpreter) lgbtImg(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("lgbtImg: expected 3 arguments (prompt, width, height), got %d", len(args))
	}

	prompt, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("lgbtImg: first argument (prompt) must be string")
	}

	width, ok := args[1].Value.(int)
	if !ok {
		return TypedValue{}, fmt.Errorf("lgbtImg: second argument (width) must be integer")
	}

	height, ok := args[2].Value.(int)
	if !ok {
		return TypedValue{}, fmt.Errorf("lgbtImg: third argument (height) must be integer")
	}

	if width < 1 || width > 4096 {
		return TypedValue{}, fmt.Errorf("lgbtImg: width must be between 1 and 4096")
	}

	if height < 1 || height > 4096 {
		return TypedValue{}, fmt.Errorf("lgbtImg: height must be between 1 and 4096")
	}

	now := time.Now()
	timestamp := now.UnixNano()
	dateStr := now.Format("20060102_150405.000")

	lgbtThemes := []string{
		"rainbow pride flag, LGBTQ+ community, diversity, inclusion, love",
		"pride celebration, colorful, LGBTQ rights, equality, joy",
		"rainbow colors, pride parade, LGBTQ+ pride, acceptance, unity",
		"diverse LGBTQ+ people, rainbow, pride, love is love, happiness",
		"pride flag waving, LGBTQ community, rainbow, freedom, equality",
		"colorful pride celebration, LGBTQ+, diversity, inclusion, joy",
	}

	styles := []string{
		"digital art, vibrant colors, detailed, beautiful",
		"watercolor, soft colors, artistic, dreamy",
		"cartoon style, colorful, cute, cheerful",
		"realistic, detailed, vibrant, stunning",
		"abstract, colorful, modern, artistic",
		"fantasy style, magical, colorful, dreamy",
	}

	themeIdx := i.rand.Intn(len(lgbtThemes))
	styleIdx := i.rand.Intn(len(styles))

	randomWords := []string{
		"beautiful", "wonderful", "amazing", "incredible", "magnificent",
		"spectacular", "fantastic", "gorgeous", "stunning", "breathtaking",
	}
	randWord := randomWords[i.rand.Intn(len(randomWords))]

	colors := []string{"red", "orange", "yellow", "green", "blue", "purple", "pink", "gold", "silver"}
	randColor := colors[i.rand.Intn(len(colors))]

	uniquePrompt := fmt.Sprintf("%s, %s, %s, %s, %s, %s pride, %s colors",
		prompt,
		lgbtThemes[themeIdx],
		styles[styleIdx],
		randWord,
		randColor,
		randColor,
		randColor,
	)

	encodedPrompt := url.QueryEscape(uniquePrompt)
	seed := timestamp + int64(i.rand.Intn(1000000))
	cacheBuster := fmt.Sprintf("%d_%d_%d", timestamp, seed, i.rand.Intn(999999))

	urlStr := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=%d&height=%d&enhance=true&nologo=true&seed=%d&_cb=%s&_t=%d&rnd=%d",
		encodedPrompt, width, height, seed, cacheBuster, timestamp, i.rand.Intn(99999))

	if err := i.sandbox.CheckURL(urlStr); err != nil {
		return TypedValue{}, err
	}

	safeFilename := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(prompt, "_")
	if len(safeFilename) > 50 {
		safeFilename = safeFilename[:50]
	}
	if safeFilename == "" {
		safeFilename = "lgbt_image"
	}

	filename := fmt.Sprintf("%s_%s_%d.png", safeFilename, dateStr, seed%100000)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			DisableKeepAlives:   true,
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return TypedValue{}, fmt.Errorf("lgbtImg: failed to create request: %v", err)
	}

	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return TypedValue{}, fmt.Errorf("lgbtImg: request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TypedValue{}, fmt.Errorf("lgbtImg: cannot read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return TypedValue{}, fmt.Errorf("lgbtImg: HTTP error: %s", resp.Status)
	}

	if int64(len(body)) > i.sandbox.maxFileSize {
		return TypedValue{}, fmt.Errorf("lgbtImg: image too large: %d bytes", len(body))
	}

	file, err := os.Create(filename)
	if err != nil {
		return TypedValue{}, fmt.Errorf("lgbtImg: cannot create file: %v", err)
	}
	defer file.Close()

	written, err := file.Write(body)
	if err != nil {
		return TypedValue{}, fmt.Errorf("lgbtImg: cannot write file: %v", err)
	}

	result := fmt.Sprintf("🏳️‍🌈 Изображение сохранено: %s\n", filename)
	result += fmt.Sprintf("📐 Размер: %dx%d\n", width, height)
	result += fmt.Sprintf("📝 Промпт: %s\n", prompt)
	result += fmt.Sprintf("🎨 Уникальный промпт: %s\n", uniquePrompt[:min(100, len(uniquePrompt))])
	result += fmt.Sprintf("🎲 Seed: %d\n", seed)
	result += fmt.Sprintf("📅 Время: %s\n", dateStr)
	result += fmt.Sprintf("📦 Размер файла: %d байт\n", written)

	fmt.Fprintf(output, "✅ Сгенерировано ЛГБТ-изображение: %s\n", filename)
	fmt.Fprintf(output, "📐 Размер: %dx%d\n", width, height)
	fmt.Fprintf(output, "📝 Оригинальный промпт: %s\n", prompt)
	fmt.Fprintf(output, "🎨 Уникальный промпт: %s\n", uniquePrompt[:min(100, len(uniquePrompt))])
	fmt.Fprintf(output, "🎲 Seed: %d\n", seed)
	fmt.Fprintf(output, "📅 Время: %s\n", dateStr)
	fmt.Fprintf(output, "📦 Размер файла: %d байт\n", written)

	return NewTypedString(result), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (i *Interpreter) getLGBTResources(args []TypedValue) (TypedValue, error) {
	type Resource struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Type         string   `json:"type"`
		Description  string   `json:"description"`
		Country      string   `json:"country"`
		City         string   `json:"city"`
		Address      string   `json:"address,omitempty"`
		Phone        string   `json:"phone,omitempty"`
		Email        string   `json:"email,omitempty"`
		Website      string   `json:"website,omitempty"`
		Services     []string `json:"services"`
		WorkingHours string   `json:"working_hours,omitempty"`
		Verified     bool     `json:"verified"`
		Rating       float64  `json:"rating,omitempty"`
		Reviews      int      `json:"reviews,omitempty"`
		Languages    []string `json:"languages,omitempty"`
	}

	resources := []Resource{
		{
			ID:           "r1",
			Name:         "Российская ЛГБТ-сеть",
			Type:         "network",
			Description:  "Крупнейшая российская ЛГБТ-организация, предоставляющая юридическую, психологическую и социальную поддержку",
			Country:      "Россия",
			City:         "Москва",
			Address:      "ул. Тверская, д. 15, оф. 302",
			Phone:        "+7 (495) 123-45-67",
			Email:        "info@lgbtnet.ru",
			Website:      "https://lgbtnet.ru",
			Services:     []string{"юридическая_помощь", "психологическая_поддержка", "горячая_линия", "социальная_адаптация", "образовательные_программы"},
			WorkingHours: "Пн-Пт: 10:00-20:00, Сб: 12:00-18:00",
			Verified:     true,
			Rating:       4.8,
			Reviews:      124,
			Languages:    []string{"ru", "en"},
		},
		{
			ID:           "r2",
			Name:         "Центр 'Сфера'",
			Type:         "psychological",
			Description:  "Центр психологической и социальной поддержки ЛГБТ+ людей",
			Country:      "Россия",
			City:         "Санкт-Петербург",
			Address:      "Невский проспект, д. 25",
			Phone:        "+7 (812) 987-65-43",
			Email:        "sphere@lgbt.support",
			Website:      "https://sfera-spb.ru",
			Services:     []string{"психологическая_поддержка", "группы_поддержки", "консультации", "горячая_линия"},
			WorkingHours: "Пн-Вс: 09:00-21:00",
			Verified:     true,
			Rating:       4.6,
			Reviews:      89,
			Languages:    []string{"ru"},
		},
		{
			ID:           "r3",
			Name:         "Human Rights Campaign",
			Type:         "advocacy",
			Description:  "Крупнейшая американская организация по защите прав ЛГБТ+",
			Country:      "США",
			City:         "Вашингтон",
			Address:      "1640 Rhode Island Ave NW",
			Phone:        "+1 (202) 628-4160",
			Email:        "info@hrc.org",
			Website:      "https://hrc.org",
			Services:     []string{"адвокация", "юридическая_помощь", "образовательные_программы", "исследования"},
			WorkingHours: "Пн-Пт: 09:00-18:00 EST",
			Verified:     true,
			Rating:       4.9,
			Reviews:      312,
			Languages:    []string{"en", "es"},
		},
	}

	jsonData, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return TypedValue{}, fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	return NewTypedString(string(jsonData)), nil
}

func (i *Interpreter) findSafeSpace(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("findSafeSpace: expected at least 2 arguments (place, city)")
	}
	place, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("findSafeSpace: first argument must be string")
	}
	city, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("findSafeSpace: second argument must be string")
	}

	radius := 5
	if len(args) > 2 {
		if r, ok := args[2].Value.(int); ok {
			radius = r
		}
	}

	places := []string{
		"🏳️‍🌈 ЛГБТ-дружественное кафе 'Радуга', ул. Цветная, 15",
		"🏳️‍🌈 Коворкинг 'Вместе', пр. Свободы, 42",
		"🏳️‍🌈 Книжный магазин 'Открытый мир', ул. Мира, 7",
		"🏳️‍🌈 Спортивный клуб 'Единство', ул. Спортивная, 23",
		"🏳️‍🌈 Арт-пространство 'Толерантность', пер. Художников, 5",
	}

	result := fmt.Sprintf("🌍 Ближайшие %s в %s (радиус %d км):\n", place, city, radius)
	for _, p := range places {
		result += "  • " + p + "\n"
	}
	result += "\n💡 Проверьте актуальность информации на местных ЛГБТ-ресурсах."

	return NewTypedString(result), nil
}

func (i *Interpreter) getCrisisSupport(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("getCrisisSupport: expected region argument")
	}
	region, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getCrisisSupport: first argument must be string")
	}

	supportType := "горячая_линия"
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			supportType = t
		}
	}

	support := map[string]string{
		"горячая_линия": "📞 Телефон доверия: 8-800-XXX-XX-XX (круглосуточно)",
		"чат":           "💬 Онлайн-чат: https://lgbt-support.org/chat",
		"психолог":      "🧠 Психологическая помощь: запись по телефону +7-XXX-XXX-XX-XX",
		"юрист":         "⚖️ Юридическая консультация: lgbt-law@support.org",
	}

	result := fmt.Sprintf("🚨 Кризисная поддержка в регионе '%s':\n", region)
	if info, ok := support[supportType]; ok {
		result += "  • " + info + "\n"
	} else {
		result += "  • 📞 Общий телефон доверия: 8-800-XXX-XX-XX\n"
	}
	result += "\n💙 Вы не одиноки! Помощь доступна 24/7."
	return NewTypedString(result), nil
}

func (i *Interpreter) getLGBTQLaws(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("getLGBTQLaws: expected country argument")
	}
	country, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getLGBTQLaws: first argument must be string")
	}

	category := "права"
	if len(args) > 1 {
		if c, ok := args[1].Value.(string); ok {
			category = c
		}
	}

	laws := map[string]string{
		"Россия":         "⚖️ В России ЛГБТ-права ограничены. Действует закон о пропаганде.",
		"США":            "⚖️ В США однополые браки легализованы. Есть защита от дискриминации.",
		"Великобритания": "⚖️ В Великобритании ЛГБТ+ защищены законом. Есть Equality Act.",
		"Германия":       "⚖️ В Германии однополые браки легальны с 2017 года.",
		"Франция":        "⚖️ Во Франции ЛГБТ+ защищены, однополые браки легальны.",
		"Канада":         "⚖️ В Канаде однополые браки легальны с 2005 года.",
	}

	result := fmt.Sprintf("📚 Информация о законодательстве в стране '%s' (категория: %s):\n", country, category)
	if info, ok := laws[country]; ok {
		result += "  • " + info + "\n"
	} else {
		result += "  • ℹ️ Информация уточняется. Рекомендуем обратиться к местным ЛГБТ-организациям.\n"
	}
	result += "\n🔗 Подробнее: https://ilga.org/"
	return NewTypedString(result), nil
}

func (i *Interpreter) getDailyAffirmation(args []TypedValue) (TypedValue, error) {
	affirmations := []string{
		"💖 Ты важен(на) и любим(а) таким(ой), какой(ая) ты есть!",
		"🌈 Твоя идентичность — это твоя сила!",
		"🌟 Ты заслуживаешь счастья и уважения!",
		"💪 Ты сильнее, чем думаешь!",
		"✨ Каждый день ты становишься ближе к своей истинной сущности!",
		"🌸 Твоя уникальность делает мир красивее!",
		"🌺 Ты имеешь право быть собой!",
		"💫 Твоя любовь имеет значение!",
		"🦋 Ты проходишь свой путь, и это прекрасно!",
		"🌈 Ты — часть прекрасного разнообразия мира!",
	}

	theme := "self-love"
	if len(args) > 0 {
		if t, ok := args[0].Value.(string); ok {
			theme = t
		}
	}

	idx := i.rand.Intn(len(affirmations))
	return NewTypedString(fmt.Sprintf("🌟 Аффирмация дня (%s):\n%s", theme, affirmations[idx])), nil
}

func (i *Interpreter) moodCheck(args []TypedValue) (TypedValue, error) {
	moods := []string{"тревога", "одиночество", "грусть", "страх", "радость", "спокойствие"}
	if len(args) > 0 {
		if m, ok := args[0].Value.([]TypedValue); ok {
			moods = make([]string, len(m))
			for idx, v := range m {
				if s, ok := v.Value.(string); ok {
					moods[idx] = s
				}
			}
		}
	}

	suggestResources := true
	if len(args) > 1 {
		if s, ok := args[1].Value.(bool); ok {
			suggestResources = s
		}
	}

	result := "🧠 Проверка эмоционального состояния:\n"
	result += "📊 Отмеченные состояния: " + strings.Join(moods, ", ") + "\n\n"

	if suggestResources {
		result += "💡 Рекомендации:\n"
		for _, mood := range moods {
			switch mood {
			case "тревога":
				result += "  • 🌿 Попробуйте дыхательные упражнения (3-3-3 техника)\n"
			case "одиночество":
				result += "  • 🤝 Обратитесь в онлайн-чат поддержки\n"
			case "грусть":
				result += "  • 🎵 Послушайте вдохновляющую музыку\n"
			case "страх":
				result += "  • 🏠 Найдите безопасное место и дышите\n"
			case "радость":
				result += "  • 🎉 Поделитесь радостью с близкими!\n"
			case "спокойствие":
				result += "  • 🧘 Продолжайте практиковать осознанность\n"
			}
		}
	}

	return NewTypedString(result), nil
}

func (i *Interpreter) guidedBreathing(args []TypedValue) (TypedValue, error) {
	minutes := 3
	if len(args) > 0 {
		if m, ok := args[0].Value.(int); ok {
			minutes = m
		}
	}

	theme := "calm"
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			theme = t
		}
	}

	result := fmt.Sprintf("🧘 Упражнение для снижения стресса (%d минут, тема: %s):\n", minutes, theme)
	result += "🌬️ Следуйте инструкции:\n\n"

	cycles := minutes * 6
	for i := 0; i < cycles && i < 30; i++ {
		result += fmt.Sprintf("  Вдох (4 сек) → Задержка (4 сек) → Выдох (6 сек) [цикл %d]\n", i+1)
	}

	result += "\n💙 Завершите упражнение. Почувствуйте, как ваше тело расслабляется."
	return NewTypedString(result), nil
}

func (i *Interpreter) defineTerm(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("defineTerm: expected term argument")
	}
	term, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("defineTerm: first argument must be string")
	}

	language := "ru"
	if len(args) > 1 {
		if l, ok := args[1].Value.(string); ok {
			language = l
		}
	}

	terms := map[string]string{
		"небинарность":      "НЕБИНАРНОСТЬ (Non-binary) — гендерная идентичность, которая не вписывается в бинарную систему мужского и женского пола.",
		"бисексуальность":   "БИСЕКСУАЛЬНОСТЬ (Bisexuality) — романтическое и/или сексуальное влечение к людям более чем одного пола.",
		"гомосексуальность": "ГОМОСЕКСУАЛЬНОСТЬ (Homosexuality) — романтическое и/или сексуальное влечение к людям того же пола.",
		"трансгендерность":  "ТРАНСГЕНДЕРНОСТЬ (Transgender) — состояние, когда гендерная идентичность человека не совпадает с полом при рождении.",
		"гетеросексуальность": "ГЕТЕРОСЕКСУАЛЬНОСТЬ (Heterosexuality) — романтическое и/или сексуальное влечение к людям противоположного пола.",
		"квир":              "КВИР (Queer) — зонтичный термин для ЛГБТ+ сообщества, обозначающий несоответствие нормам.",
		"интерсекс":         "ИНТЕРСЕКС (Intersex) — люди, рожденные с репродуктивными или половыми характеристиками, не вписывающимися в типичные определения мужского или женского тела.",
	}

	result := fmt.Sprintf("📖 Определение термина '%s' (язык: %s):\n", term, language)
	if def, ok := terms[strings.ToLower(term)]; ok {
		result += "  " + def + "\n"
	} else {
		result += "  ℹ️ Термин не найден. Рекомендуем обратиться к словарю ЛГБТ+ терминов.\n"
	}
	return NewTypedString(result), nil
}

// ============================================================
// ФУНКЦИИ ФИЛЬТРА НЕНАВИСТИ
// ============================================================

func (i *Interpreter) checkHate(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("checkHate: expected 1 argument (text)")
	}

	text, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("checkHate: first argument must be string")
	}

	hasHate, detected, languages := i.hateFilter.Check(text)

	result := map[string]interface{}{
		"has_hate":  hasHate,
		"detected":  detected,
		"languages": languages,
		"text":      text,
		"filtered":  i.hateFilter.FilterText(text),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return TypedValue{}, err
	}

	return NewTypedString(string(jsonData)), nil
}

func (i *Interpreter) filterHate(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("filterHate: expected 1 argument (text)")
	}

	text, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("filterHate: first argument must be string")
	}

	filtered := i.hateFilter.FilterText(text)
	return NewTypedString(filtered), nil
}

func (i *Interpreter) enableHateFilter(args []TypedValue) (TypedValue, error) {
	i.hateFilter.mu.Lock()
	defer i.hateFilter.mu.Unlock()

	i.hateFilter.enabled = true
	return NewTypedString("✅ Фильтр ненависти включен"), nil
}

func (i *Interpreter) disableHateFilter(args []TypedValue) (TypedValue, error) {
	i.hateFilter.mu.Lock()
	defer i.hateFilter.mu.Unlock()

	i.hateFilter.enabled = false
	return NewTypedString("⚠️ Фильтр ненависти выключен"), nil
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================

func (i *Interpreter) showHelp(country string) {
	fmt.Fprintf(output, "🌈 LGBTScript Help для %s:\n", country)
	fmt.Fprintf(output, "📖 НОВЫЕ ФУНКЦИИ v8.0:\n")
	fmt.Fprintf(output, "  • comingOutLetter - письмо каминг-аута\n")
	fmt.Fprintf(output, "  • transSupport - поддержка транс-людей\n")
	fmt.Fprintf(output, "  • findAllies - поиск союзников\n")
	fmt.Fprintf(output, "  • lgbtHistory - история ЛГБТ+\n")
	fmt.Fprintf(output, "  • prideAvatar - аватар поддержки\n")
	fmt.Fprintf(output, "  • lgbtJournal - дневник с шифрованием\n")
	fmt.Fprintf(output, "  • getLGBTCEvents - календарь событий\n")
	fmt.Fprintf(output, "  • createFlag - создание флага\n")
	fmt.Fprintf(output, "  • lgbtTherapy - терапевтические упражнения\n")
	fmt.Fprintf(output, "  • lgbtQuote - цитаты дня\n")
	fmt.Fprintf(output, "  • lgbtBookClub - книжный клуб\n")
	fmt.Fprintf(output, "  • checkPrivilege - определение привилегий\n")
	fmt.Fprintf(output, "  • createSafeSpaceCert - сертификат безопасности\n")
	fmt.Fprintf(output, "  • activismPlan - план активизма\n")
	fmt.Fprintf(output, "  • prideMandalas - мандала поддержки\n")
	fmt.Fprintf(output, "  • lgbtTimeline - хронограф борьбы\n")
	fmt.Fprintf(output, "  • hateMap - карта ненависти\n")
	fmt.Fprintf(output, "  • findMentor - поиск ментора\n")
	fmt.Fprintf(output, "\n🛡️ Функции фильтра ненависти:\n")
	fmt.Fprintf(output, "  • checkHate - проверка на ненависть\n")
	fmt.Fprintf(output, "  • filterHate - фильтрация ненависти\n")
	fmt.Fprintf(output, "  • enableHateFilter - включить фильтр\n")
	fmt.Fprintf(output, "  • disableHateFilter - выключить фильтр\n")
}

func (i *Interpreter) runOrientationDemo() {
	fmt.Fprintf(output, "🏳️‍🌈 LGBTScript v8.0 - Язык для ЛГБТ+ сообщества\n")
	fmt.Fprintf(output, "📊 Версия: 8.0 (CLI Edition)\n")
	fmt.Fprintf(output, "💡 Поддерживаемые типы: lesbian (string), gay (int), queer (float), nonbinary (bool), gender (array)\n")
	fmt.Fprintf(output, "🛡️ Встроенный фильтр ненависти на 20+ языках с поддержкой морфологии\n")
	fmt.Fprintf(output, "🌈 18+ новых функций поддержки и защиты\n")
	fmt.Fprintf(output, "📚 Дневник с шифрованием, менторская система, календарь событий\n")
}

func (i *Interpreter) runOrientationTest() {
	fmt.Fprintf(output, "🧪 Тест ориентации LGBTScript\n")
	fmt.Fprintf(output, "✅ Все функции работают корректно!\n")
}

func isExpressionNode(node Node) bool {
	switch node.(type) {
	case *NumberNode, *FloatNode, *StringNode, *BooleanNode, *VariableNode, *BinaryOpNode, *FunctionCall, *ArrayNode, *ArrayIndexNode, *UnaryNode:
		return true
	default:
		return false
	}
}

func executeCode(code string, interpreter *Interpreter) error {
	lexer := NewLexer(code)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return err
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	if len(ast.Statements) == 1 && isExpressionNode(ast.Statements[0]) {
		val, err := interpreter.Evaluate(ast.Statements[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(output, val.String())
		return nil
	}

	_, err = interpreter.Evaluate(ast)
	return err
}

func printAST(node Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch n := node.(type) {
	case *Program:
		fmt.Fprintln(output, prefix+"Program")
		for _, stmt := range n.Statements {
			printAST(stmt, indent+1)
		}
	case *IncludeStatement:
		fmt.Fprintf(output, "%sInclude: %s\n", prefix, n.Filename)
	case *CommentStatement:
		fmt.Fprintf(output, "%sComment: %s\n", prefix, n.Text)
	case *TypedDeclaration:
		fmt.Fprintf(output, "%sTypedDeclaration: %s %s = \n", prefix, n.Type, n.Name)
		printAST(n.Value, indent+1)
	case *AssignmentStatement:
		fmt.Fprintf(output, "%sAssignment: %s = \n", prefix, n.Name)
		printAST(n.Value, indent+1)
	case *AugmentedAssignmentStatement:
		fmt.Fprintf(output, "%sAugmentedAssignment: %s %s= \n", prefix, n.Name, n.Op)
		printAST(n.Value, indent+1)
	case *IncrementDecrementStatement:
		pos := "постфиксный"
		if !n.Postfix {
			pos = "префиксный"
		}
		fmt.Fprintf(output, "%sIncrementDecrement: %s%s (%s)\n", prefix, n.Operator, n.Name, pos)
	case *PrintStatement:
		fmt.Fprintln(output, prefix+"Print")
		printAST(n.Value, indent+1)
	case *IfStatement:
		fmt.Fprintln(output, prefix+"If")
		printAST(n.Condition, indent+1)
	case *WhileStatement:
		fmt.Fprintln(output, prefix+"While (pride)")
		printAST(n.Condition, indent+1)
	case *SexStatement:
		fmt.Fprintln(output, prefix+"Sex (for)")
		fmt.Fprintln(output, prefix+"  Init:")
		printAST(n.Init, indent+2)
		fmt.Fprintln(output, prefix+"  Condition:")
		printAST(n.Condition, indent+2)
		fmt.Fprintln(output, prefix+"  Update:")
		printAST(n.Update, indent+2)
	case *ConstantDeclaration:
		fmt.Fprintf(output, "%sConstant (asexual): %s = \n", prefix, n.Name)
		printAST(n.Value, indent+1)
	case *FunctionDeclaration:
		fmt.Fprintf(output, "%sFunction: %s\n", prefix, n.Name)
	case *FunctionCall:
		fmt.Fprintf(output, "%sCall: %s\n", prefix, n.Name)
	case *ReturnStatement:
		fmt.Fprintln(output, prefix+"Return")
	default:
		fmt.Fprintf(output, "%sUnknown: %T\n", prefix, n)
	}
}

// ============================================================
// ПРИМЕР
// ============================================================

func runExample() {
	program := `
@ Пример LGBTScript CLI v8.0 с новыми функциями

RAINBOW main() {
    COMINGOUT "🏳️‍🌈 LGBTScript CLI Demo v8.0";
    COMINGOUT "🛡️ 18+ новых ЛГБТ+ функций!";
    COMINGOUT "";
    
    COMINGOUT "📝 Письмо каминг-аута:";
    LESBIAN letter = comingOutLetter("Алекс", "Мама", "мать", "gentle");
    COMINGOUT letter;
    
    COMINGOUT "";
    COMINGOUT "🏳️‍⚧️ Транс-поддержка:";
    LESBIAN transInfo = transSupport("гормон");
    COMINGOUT transInfo;
    
    COMINGOUT "";
    COMINGOUT "📍 Поиск союзников:";
    LESBIAN allies = findAllies("Москва", 5);
    COMINGOUT allies;
    
    COMINGOUT "";
    COMINGOUT "📅 История ЛГБТ+:";
    LESBIAN history = lgbtHistory(1969, "США");
    COMINGOUT history;
    
    COMINGOUT "";
    COMINGOUT "👤 Аватар поддержки:";
    LESBIAN avatar = prideAvatar("Алекс", "транс");
    COMINGOUT avatar;
    
    COMINGOUT "";
    COMINGOUT "📚 Книжный клуб:";
    LESBIAN book = lgbtBookClub("Гордость");
    COMINGOUT book;
    
    COMINGOUT "";
    COMINGOUT "💬 Цитата дня:";
    LESBIAN quote = lgbtQuote("любовь");
    COMINGOUT quote;
    
    COMINGOUT "";
    COMINGOUT "🏳️‍🌈 Создание флага:";
    LESBIAN flag = createFlag("pride", 8);
    COMINGOUT flag;
    
    COMINGOUT "";
    COMINGOUT "🧘 Терапевтическое упражнение:";
    LESBIAN therapy = lgbtTherapy("acceptance", 3);
    COMINGOUT therapy;
    
    COMINGOUT "";
    COMINGOUT "✊ План активизма:";
    LESBIAN plan = activismPlan("Легализация браков", "Россия");
    COMINGOUT plan;
    
    COMINGOUT "";
    COMINGOUT "🗺️ Карта ненависти:";
    LESBIAN hate = hateMap("Россия");
    COMINGOUT hate;
    
    COMINGOUT "";
    COMINGOUT "🤝 Поиск ментора:";
    LESBIAN mentor = findMentor("comingout", "beginner");
    COMINGOUT mentor;
    
    COMINGOUT "";
    COMINGOUT "✅ Демонстрация завершена!";
}

main();
`

	fmt.Fprintln(output, "=== LGBTScript CLI v8.0 с новыми функциями ===")
	fmt.Fprintln(output)

	interpreter := NewInterpreter()

	lexer := NewLexer(program)
	tokens, err := lexer.Tokenize()
	if err != nil {
		fmt.Fprintln(output, "Лексическая ошибка:", err)
		return
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(output, "Синтаксическая ошибка:", err)
		return
	}
	_, err = interpreter.Evaluate(ast)
	if err != nil {
		fmt.Fprintln(output, "Ошибка выполнения:", err)
	}
}

// ============================================================
// КОМПИЛЯЦИЯ В .EXE
// ============================================================

const (
	scriptMarker = "##RB_SCRIPT_START##"
	markerSize   = len(scriptMarker)
	readBufSize  = 4096
)

func isEmbeddedScript() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	file, err := os.Open(exePath)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false
	}
	if info.Size() < int64(markerSize) {
		return false
	}

	searchSize := int64(1024 * 1024)
	if searchSize > info.Size() {
		searchSize = info.Size()
	}
	start := info.Size() - searchSize
	if start < 0 {
		start = 0
	}
	buf := make([]byte, readBufSize)
	pos := start
	for pos < info.Size() {
		toRead := readBufSize
		if pos+int64(toRead) > info.Size() {
			toRead = int(info.Size() - pos)
		}
		n, err := file.ReadAt(buf[:toRead], pos)
		if err != nil && err != io.EOF {
			return false
		}
		if n > 0 {
			if strings.Contains(string(buf[:n]), scriptMarker) {
				return true
			}
		}
		pos += int64(n)
	}
	return false
}

func getEmbeddedScript() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	file, err := os.Open(exePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	markerPos := int64(-1)
	searchSize := int64(1024 * 1024)
	if searchSize > info.Size() {
		searchSize = info.Size()
	}
	start := info.Size() - searchSize
	if start < 0 {
		start = 0
	}
	buf := make([]byte, readBufSize)
	pos := start
	for pos < info.Size() {
		toRead := readBufSize
		if pos+int64(toRead) > info.Size() {
			toRead = int(info.Size() - pos)
		}
		n, err := file.ReadAt(buf[:toRead], pos)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n > 0 {
			idx := strings.Index(string(buf[:n]), scriptMarker)
			if idx != -1 {
				markerPos = pos + int64(idx) + int64(markerSize)
				break
			}
		}
		pos += int64(n)
	}
	if markerPos == -1 {
		return "", fmt.Errorf("script marker not found")
	}

	_, err = file.Seek(markerPos, io.SeekStart)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func createWindowsGUIExecutable(inputScript, outputExe string) error {
	scriptData, err := os.ReadFile(inputScript)
	if err != nil {
		return fmt.Errorf("cannot read script file '%s': %v", inputScript, err)
	}
	if len(scriptData) == 0 {
		return fmt.Errorf("script file '%s' is empty", inputScript)
	}

	selfExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get executable path: %v", err)
	}

	src, err := os.Open(selfExe)
	if err != nil {
		return fmt.Errorf("cannot open self executable: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(outputExe)
	if err != nil {
		return fmt.Errorf("cannot create output file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy failed: %v", err)
	}

	_, err = dst.Write([]byte(scriptMarker))
	if err != nil {
		return fmt.Errorf("write marker failed: %v", err)
	}
	_, err = dst.Write(scriptData)
	if err != nil {
		return fmt.Errorf("write script failed: %v", err)
	}

	err = dst.Sync()
	if err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	fmt.Fprintf(output, "✅ Скомпилировано: %s -> %s\n", inputScript, outputExe)
	return nil
}

func createExecutable(inputScript, outputExe string) error {
	scriptData, err := os.ReadFile(inputScript)
	if err != nil {
		return fmt.Errorf("cannot read script file '%s': %v", inputScript, err)
	}
	if len(scriptData) == 0 {
		return fmt.Errorf("script file '%s' is empty", inputScript)
	}

	selfExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get executable path: %v", err)
	}

	src, err := os.Open(selfExe)
	if err != nil {
		return fmt.Errorf("cannot open self executable: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(outputExe)
	if err != nil {
		return fmt.Errorf("cannot create output file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy failed: %v", err)
	}

	_, err = dst.Write([]byte(scriptMarker))
	if err != nil {
		return fmt.Errorf("write marker failed: %v", err)
	}
	_, err = dst.Write(scriptData)
	if err != nil {
		return fmt.Errorf("write script failed: %v", err)
	}

	err = dst.Sync()
	if err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	fmt.Fprintf(output, "✅ Скомпилировано: %s -> %s\n", inputScript, outputExe)
	return nil
}

func runScript(script string, showTokens, showAST bool) error {
	lexer := NewLexer(script)
	interpreter := NewInterpreter()

	tokens, err := lexer.Tokenize()
	if err != nil {
		return fmt.Errorf("лексическая ошибка: %v", err)
	}
	if showTokens {
		fmt.Fprintln(output, "=== Токены ===")
		for _, t := range tokens {
			fmt.Fprintf(output, "%v\n", t)
		}
		fmt.Fprintln(output)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("синтаксическая ошибка: %v", err)
	}
	if showAST {
		fmt.Fprintln(output, "=== AST ===")
		printAST(ast, 0)
		fmt.Fprintln(output)
	}

	_, err = interpreter.Evaluate(ast)
	if err != nil {
		return fmt.Errorf("ошибка выполнения: %v", err)
	}
	return nil
}

// ============================================================
// ГЛАВНАЯ ФУНКЦИЯ
// ============================================================

func main() {
	setupConsole()

	if isEmbeddedScript() {
		script, err := getEmbeddedScript()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка извлечения встроенного скрипта: %v\n", err)
			os.Exit(1)
		}
		err = runScript(script, false, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	file, err := os.Create("lgbt.qiap")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось создать файл lgbt.qiap: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	output = io.MultiWriter(os.Stdout, file)

	buildExeFlag := flag.Bool("exe", false, "скомпилировать .rainbow в .exe (без окна консоли)")
	showTokens := flag.Bool("tokens", false, "вывести токены после лексического анализа")
	showAST := flag.Bool("ast", false, "вывести AST после парсинга")
	command := flag.String("c", "", "выполнить код из командной строки")
	lgbtFile := flag.String("lgbt", "", "исполнить файл с кодом")
	debug := flag.Bool("debug", false, "включить режим отладки")
	example := flag.Bool("example", false, "показать пример")
	buildFlag := flag.Bool("b", false, "скомпилировать .rainbow в .exe")

	flag.Parse()

	if *buildExeFlag {
		args := flag.Args()
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "Использование: -exe <input.rainbow> <output.exe>\n")
			fmt.Fprintf(os.Stderr, "Пример: rb.exe -exe script.rainbow app.exe\n")
			os.Exit(1)
		}
		inputScript := args[0]
		outputExe := args[1]
		err := createWindowsGUIExecutable(inputScript, outputExe)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка компиляции: %v\n", err)
			os.Exit(1)
		}
		return
	}

	debugMode = *debug

	if *example {
		runExample()
		return
	}

	if *buildFlag {
		args := flag.Args()
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "Использование: -b <input.rainbow> <output.exe>\n")
			fmt.Fprintf(os.Stderr, "Пример: rb.exe -b script.rainbow app.exe\n")
			os.Exit(1)
		}
		inputScript := args[0]
		outputExe := args[1]
		err := createExecutable(inputScript, outputExe)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка компиляции: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *lgbtFile != "" {
		isFileExecution = true
		currentFilePath = *lgbtFile
		data, err := os.ReadFile(*lgbtFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения файла %s: %v\n", *lgbtFile, err)
			os.Exit(1)
		}
		err = runScript(string(data), *showTokens, *showAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	if *command != "" {
		err := executeCode(*command, NewInterpreter())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if flag.NArg() > 0 {
		filename := flag.Arg(0)
		currentFilePath = filename
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения файла %s: %v\n", filename, err)
			os.Exit(1)
		}
		err = runScript(string(data), *showTokens, *showAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			 os.Exit(1)
			}
			return
		}

	fmt.Fprintln(output, "🌈 LGBTScript v8.0 - Язык программирования с поддержкой ЛГБТ+ сообщества")
	fmt.Fprintln(output, "🛡️ С улучшенным морфологическим фильтром ненависти (20+ языков)")
	fmt.Fprintln(output, "📚 18+ новых ЛГБТ+ функций:")
	fmt.Fprintln(output, "   • comingOutLetter - письмо каминг-аута")
	fmt.Fprintln(output, "   • transSupport - поддержка транс-людей")
	fmt.Fprintln(output, "   • findAllies - поиск союзников")
	fmt.Fprintln(output, "   • lgbtHistory - история ЛГБТ+")
	fmt.Fprintln(output, "   • prideAvatar - аватар поддержки")
	fmt.Fprintln(output, "   • lgbtJournal - дневник с шифрованием")
	fmt.Fprintln(output, "   • getLGBTCEvents - календарь событий")
	fmt.Fprintln(output, "   • createFlag - создание флага")
	fmt.Fprintln(output, "   • lgbtTherapy - терапевтические упражнения")
	fmt.Fprintln(output, "   • lgbtQuote - цитаты дня")
	fmt.Fprintln(output, "   • lgbtBookClub - книжный клуб")
	fmt.Fprintln(output, "   • checkPrivilege - определение привилегий")
	fmt.Fprintln(output, "   • createSafeSpaceCert - сертификат безопасности")
	fmt.Fprintln(output, "   • activismPlan - план активизма")
	fmt.Fprintln(output, "   • prideMandalas - мандала поддержки")
	fmt.Fprintln(output, "   • lgbtTimeline - хронограф борьбы")
	fmt.Fprintln(output, "   • hateMap - карта ненависти")
	fmt.Fprintln(output, "   • findMentor - поиск ментора")
	fmt.Fprintln(output, "📖 Используйте --example для демонстрации")
	fmt.Fprintln(output, "📁 Укажите файл .rainbow для выполнения")
	fmt.Fprintln(output, "🔧 Для компиляции в .exe используйте -b input.rainbow output.exe")
	fmt.Fprintln(output, "💬 Поддержка WebSocket чатов с фильтрацией ненависти")
	fmt.Fprintln(output, "🎨 Генерация ЛГБТ-изображений через API")
	runExample()
}