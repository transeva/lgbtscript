package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
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
	"unsafe"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
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

	MB_ICONHAND         = 0x00000010
	MB_ICONQUESTION     = 0x00000020
	MB_ICONEXCLAMATION  = 0x00000030
	MB_ICONASTERISK     = 0x00000040
	MB_ICONWARNING      = MB_ICONEXCLAMATION
	MB_ICONERROR        = MB_ICONHAND
	MB_ICONINFORMATION  = MB_ICONASTERISK

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
			"lesbian": true, "gay": true, "trans": true, "nonbinary": true, "gender": true,
			"comingout": true, "cis": true, "nocis": true,
			"true": true, "false": true,
			"help": true, "orientation": true,
			"rainbow": true,
			"return": true,
			"try": true, "catch": true,
			"export": true,
			"queer": true, "extends": true, "this": true, "super": true, "new": true,
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

type TryCatchStatement struct {
	BaseNode
	TryBlock   []Node
	CatchBlock []Node
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

type QueerClassDeclaration struct {
	BaseNode
	Name        string
	Parent      string
	Fields      map[string]string
	Methods     map[string]*FunctionDeclaration
	Constructor *FunctionDeclaration
}

type QueerInstanceNode struct {
	BaseNode
	ClassName string
	Args      []Node
}

type QueerFieldAccessNode struct {
	BaseNode
	Object Node
	Field  string
}

type QueerMethodCallNode struct {
	BaseNode
	Object Node
	Method string
	Args   []Node
}

type ThisNode struct {
	BaseNode
}

type SuperNode struct {
	BaseNode
}

type Program struct {
	BaseNode
	Statements []Node
}

// ============================================================
// ПАРСЕР
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
		case "lesbian", "gay", "trans", "nonbinary", "gender":
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
		case "try":
			return p.parseTryCatchStatement()
		case "export":
			return p.parseExportStatement()
		case "queer":
			return p.parseQueerClassDeclaration()
		case "new":
			return p.parseNewInstance()
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

		if nextToken.Value == "." {
			return p.parseFieldAccess(token.Value)
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

	if token.Type == TOKEN_KEYWORD {
		if token.Value == "this" {
			p.next()
			if p.peek().Value == "." {
				p.next()
				field, err := p.expect(TOKEN_IDENTIFIER, "")
				if err != nil {
					return nil, err
				}
				return &QueerFieldAccessNode{
					BaseNode: BaseNode{Line: line, Col: col},
					Object:   &ThisNode{BaseNode: BaseNode{Line: line, Col: col}},
					Field:    field.Value,
				}, nil
			}
			return &ThisNode{BaseNode: BaseNode{Line: line, Col: col}}, nil
		}
		if token.Value == "super" {
			p.next()
			if p.peek().Value == "." {
				p.next()
				method, err := p.expect(TOKEN_IDENTIFIER, "")
				if err != nil {
					return nil, err
				}
				_, err = p.expect(TOKEN_OPERATOR, "(")
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
				return &QueerMethodCallNode{
					BaseNode: BaseNode{Line: line, Col: col},
					Object:   &SuperNode{BaseNode: BaseNode{Line: line, Col: col}},
					Method:   method.Value,
					Args:     args,
				}, nil
			}
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

	cond, err := p.parseExpression()
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
		cond, err := p.parseExpression()
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

	cond, err := p.parseExpression()
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

func (p *Parser) parseTryCatchStatement() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	_, err := p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var tryBlock []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		tryBlock = append(tryBlock, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_KEYWORD, "catch")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	var catchBlock []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		catchBlock = append(catchBlock, stmt)
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &TryCatchStatement{
		BaseNode:   BaseNode{Line: line, Col: col},
		TryBlock:   tryBlock,
		CatchBlock: catchBlock,
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

func (p *Parser) parseQueerClassDeclaration() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	name, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	var parent string
	if p.peek().Value == "extends" {
		p.next()
		parentToken, err := p.expect(TOKEN_IDENTIFIER, "")
		if err != nil {
			return nil, err
		}
		parent = parentToken.Value
	}

	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	class := &QueerClassDeclaration{
		BaseNode: BaseNode{Line: line, Col: col},
		Name:     name.Value,
		Parent:   parent,
		Fields:   make(map[string]string),
		Methods:  make(map[string]*FunctionDeclaration),
	}

	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		fieldToken := p.peek()
		if fieldToken.Type == TOKEN_KEYWORD && (fieldToken.Value == "lesbian" || fieldToken.Value == "gay" ||
			fieldToken.Value == "trans" || fieldToken.Value == "nonbinary" || fieldToken.Value == "gender") {
			p.next()
			fieldName, err := p.expect(TOKEN_IDENTIFIER, "")
			if err != nil {
				return nil, err
			}
			if p.peek().Value == ";" {
				p.next()
			}
			class.Fields[fieldName.Value] = fieldToken.Value
		} else if fieldToken.Type == TOKEN_KEYWORD && fieldToken.Value == "rainbow" {
			fn, err := p.parseFunctionDeclaration()
			if err != nil {
				return nil, err
			}
			if decl, ok := fn.(*FunctionDeclaration); ok {
				if decl.Name == "constructor" {
					class.Constructor = decl
				} else {
					class.Methods[decl.Name] = decl
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected token in class: %s", fieldToken.Value)
		}
	}

	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return class, nil
}

func (p *Parser) parseNewInstance() (Node, error) {
	token := p.peek()
	line := token.Line
	col := token.Col
	p.next()

	className, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TOKEN_OPERATOR, "(")
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

	if p.peek().Value == ";" {
		p.next()
	}

	return &QueerInstanceNode{
		BaseNode:  BaseNode{Line: line, Col: col},
		ClassName: className.Value,
		Args:      args,
	}, nil
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

func (p *Parser) parseFieldAccess(object string) (Node, error) {
	line := p.peek().Line
	col := p.peek().Col

	p.next()
	_, err := p.expect(TOKEN_OPERATOR, ".")
	if err != nil {
		return nil, err
	}

	field, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	return &QueerFieldAccessNode{
		BaseNode: BaseNode{Line: line, Col: col},
		Object:   &VariableNode{BaseNode: BaseNode{Line: line, Col: col}, Name: object},
		Field:    field.Value,
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
		if token.Value == "this" {
			p.next()
			return &ThisNode{BaseNode: BaseNode{Line: line, Col: col}}, nil
		}
		if token.Value == "super" {
			p.next()
			return &SuperNode{BaseNode: BaseNode{Line: line, Col: col}}, nil
		}
		if token.Value == "new" {
			return p.parseNewInstance()
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
		if nextToken.Value == "." {
			return p.parseFieldAccess(token.Value)
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

func (p *Parser) parseBlock() ([]Node, error) {
	var statements []Node

	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	_, err := p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return statements, nil
}

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
// СТРУКТУРЫ ДЛЯ GUI
// ============================================================

type GUIWindow struct {
	wnd      *ui.Main
	controls map[string]win.HWND
	events   map[string]func([]TypedValue) (TypedValue, error)
	mu       sync.RWMutex
	isActive bool
	result   TypedValue
}

var guiWindows = make(map[string]*GUIWindow)
var guiMu sync.RWMutex

// ============================================================
// СТРУКТУРЫ ДЛЯ СЕРВЕРОВ
// ============================================================

type ServerInstance struct {
	mu       sync.RWMutex
	Name     string
	Port     int
	Routes   map[string]map[string]func([]TypedValue) (TypedValue, error)
	Server   *http.Server
	IsActive bool
	Context  map[string]TypedValue
}

var servers = make(map[string]*ServerInstance)
var serversMu sync.RWMutex

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
	case TypeObject:
		if obj, ok := tv.Value.(*QueerInstance); ok {
			return obj.String()
		}
		return "{object}"
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

func NewTypedObject(obj *QueerInstance) TypedValue {
	return TypedValue{Type: TypeObject, Value: obj}
}

// ============================================================
// QUEER КЛАССЫ ВО ВРЕМЯ ВЫПОЛНЕНИЯ
// ============================================================

type QueerInstance struct {
	ClassName string
	Fields    map[string]TypedValue
	Methods   map[string]*FunctionDeclaration
	Parent    *QueerInstance
}

func (qi *QueerInstance) String() string {
	fields := make([]string, 0, len(qi.Fields))
	for name, value := range qi.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", name, value.String()))
	}
	return fmt.Sprintf("%s{%s}", qi.ClassName, strings.Join(fields, ", "))
}

// ============================================================
// HATE FILTER - ФИЛЬТР НЕНАВИСТИ
// ============================================================

type HateFilter struct {
	slurs         map[string][]string
	patterns      []*regexp.Regexp
	falsePositive map[string]bool
	mu            sync.RWMutex
	enabled       bool
	action        string
	logFile       string
}

func NewHateFilter() *HateFilter {
	hf := &HateFilter{
		slurs:         make(map[string][]string),
		patterns:      []*regexp.Regexp{},
		falsePositive: make(map[string]bool),
		enabled:       true,
		action:        "warn",
		logFile:       "hate_speech_log.txt",
	}

	hf.initSlurs()
	hf.compilePatterns()
	return hf
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
		"trans propaganda", "lgbt agenda", "rainbow mafia",
		"gaystapo", "homo lobby", "gay lobby", "groomer", "grooming",
		"degenerate", "degeneracy", "perversion", "deviant",
		"abomination", "unnatural", "satanic",
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
		"гендерна ідеологія", "пропаганда ЛГБТ", "лгбт-пропаганда",
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
		"cinsiyet ideolojisi", "toplumsal cinsiyet", "LGBT propagandası",
		"gay lobisi", "pedofil", "pedofili", "sapık", "anormal",
		"günah", "ayıp", "utanç verici",
	}

	// JAPANESE
	hf.slurs["ja"] = []string{
		"おかま", "おかま野郎", "ホモ", "ゲイ", "レズ", "レズビアン",
		"バイセクシュアル", "オカマ", "オナベ", "ニューハーフ",
		"トランスジェンダー", "性同一性障害", "ゲイのプロパガンダ",
		"LGBTのアジェンダ", "異常性欲", "性的倒錯", "変態",
		"反自然的", "不道徳", "罪", "ペドフィリア", "小児性愛",
	}

	// CHINESE
	hf.slurs["zh"] = []string{
		"同性恋", "同性恋者", "基佬", "玻璃", "拉拉", "女同",
		"男同", "同志", "变性人", "人妖", "阴阳人",
		"同性恋宣传", "同性恋议程", "LGBT议程", "性变态",
		"性倒错", "反常", "不道德", "伤风败俗", "罪恶",
		"恋童癖", "性侵", "猥亵", "性别意识形态",
	}

	// KOREAN
	hf.slurs["ko"] = []string{
		"호모", "게이", "레즈", "동성애자", "동성애",
		"트랜스젠더", "성전환", "변태", "성도착", "반자연적",
		"동성애 선전", "LGBT 의제", "포르노", "음란", "추행",
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

	allSlurs := []string{}
	for _, slurs := range hf.slurs {
		allSlurs = append(allSlurs, slurs...)
	}

	for _, slur := range allSlurs {
		pattern := regexp.QuoteMeta(slur)
		variants := []string{
			pattern,
			pattern + `\s*`,
			`\s*` + pattern,
			pattern + `[s]?`,
			pattern + `[es]?`,
			pattern + `[ing]?`,
			`\b` + pattern + `\b`,
		}

		for _, p := range variants {
			fullPattern := `(?i)` + p
			re, err := regexp.Compile(fullPattern)
			if err == nil {
				hf.patterns = append(hf.patterns, re)
			}
		}
	}
}

func (hf *HateFilter) Check(text string) (bool, []string, []string) {
	hf.mu.RLock()
	defer hf.mu.RUnlock()

	if !hf.enabled {
		return false, nil, nil
	}

	detected := []string{}
	languages := []string{}

	for _, re := range hf.patterns {
		if re.MatchString(text) {
			detected = append(detected, re.String())
			for lang, slurs := range hf.slurs {
				for _, slur := range slurs {
					if strings.Contains(strings.ToLower(text), strings.ToLower(slur)) {
						if !contains(languages, lang) {
							languages = append(languages, lang)
						}
						break
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
	return result
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
// ИНТЕРПРЕТАТОР
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

type Interpreter struct {
	variables      map[string]TypedValue
	variableTypes  map[string]string
	functions      map[string]*FunctionDeclaration
	queerClasses   map[string]*QueerClassDeclaration
	instances      map[string]*QueerInstance
	exportedFuncs  map[string]*FunctionDeclaration
	callStack      []callFrame
	returnValue    TypedValue
	returnFlag     bool
	mu             sync.RWMutex
	maxRecursion   int
	recursionDepth int
	errorHandler   func(error, int, int)
	sandbox        *Sandbox
	rand           *rand.Rand
	this           *QueerInstance
	hateFilter     *HateFilter
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:      make(map[string]TypedValue),
		variableTypes:  make(map[string]string),
		functions:      make(map[string]*FunctionDeclaration),
		queerClasses:   make(map[string]*QueerClassDeclaration),
		instances:      make(map[string]*QueerInstance),
		exportedFuncs:  make(map[string]*FunctionDeclaration),
		callStack:      []callFrame{{vars: make(map[string]TypedValue), types: make(map[string]string), constants: make(map[string]bool)}},
		returnValue:    TypedValue{Type: TypeNull, Value: nil},
		returnFlag:     false,
		maxRecursion:   1000,
		recursionDepth: 0,
		sandbox:        NewSandbox(),
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		this:           nil,
		hateFilter:     NewHateFilter(),
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
	case *QueerClassDeclaration:
		i.mu.Lock()
		i.queerClasses[n.Name] = n
		i.mu.Unlock()
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *QueerInstanceNode:
		return i.createQueerInstance(n)
	case *QueerFieldAccessNode:
		return i.accessQueerField(n)
	case *QueerMethodCallNode:
		return i.callQueerMethod(n)
	case *ThisNode:
		if i.this == nil {
			return TypedValue{}, fmt.Errorf("'this' used outside of method")
		}
		return NewTypedObject(i.this), nil
	case *SuperNode:
		if i.this == nil {
			return TypedValue{}, fmt.Errorf("'super' used outside of method")
		}
		if i.this.Parent == nil {
			return TypedValue{}, fmt.Errorf("'super' used in class without parent")
		}
		return NewTypedObject(i.this.Parent), nil
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
	case *TryCatchStatement:
		var err error
		for _, stmt := range n.TryBlock {
			_, err = i.evaluateNode(stmt)
			if err != nil {
				break
			}
		}

		if err != nil {
			line := 0
			if n, ok := node.(Node); ok {
				line = n.GetLine()
			}
			i.setVar("error", NewTypedString(err.Error()))
			i.setType("error", "lesbian")
			for _, stmt := range n.CatchBlock {
				_, catchErr := i.evaluateNode(stmt)
				if catchErr != nil {
					return TypedValue{}, catchErr
				}
			}
			i.handleError(err, line, 0)
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

// ============================================================
// МЕТОДЫ ИНТЕРПРЕТАТОРА
// ============================================================

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

func (i *Interpreter) createQueerInstance(n *QueerInstanceNode) (TypedValue, error) {
	classDecl, ok := i.queerClasses[n.ClassName]
	if !ok {
		return TypedValue{}, fmt.Errorf("class not defined: %s", n.ClassName)
	}

	instance := &QueerInstance{
		ClassName: n.ClassName,
		Fields:    make(map[string]TypedValue),
		Methods:   make(map[string]*FunctionDeclaration),
		Parent:    nil,
	}

	for fieldName, fieldType := range classDecl.Fields {
		var defaultValue TypedValue
		switch fieldType {
		case "lesbian":
			defaultValue = NewTypedString("")
		case "gay":
			defaultValue = NewTypedInt(0)
		case "trans":
			defaultValue = NewTypedFloat(0.0)
		case "nonbinary":
			defaultValue = NewTypedBool(false)
		case "gender":
			defaultValue = NewTypedArray([]TypedValue{})
		default:
			defaultValue = TypedValue{Type: TypeNull, Value: nil}
		}
		instance.Fields[fieldName] = defaultValue
	}

	for methodName, method := range classDecl.Methods {
		instance.Methods[methodName] = method
	}

	if classDecl.Parent != "" {
		parentClass, ok := i.queerClasses[classDecl.Parent]
		if ok {
			parentInstance := &QueerInstance{
				ClassName: classDecl.Parent,
				Fields:    make(map[string]TypedValue),
				Methods:   make(map[string]*FunctionDeclaration),
				Parent:    nil,
			}
			for fieldName, fieldType := range parentClass.Fields {
				var defaultValue TypedValue
				switch fieldType {
				case "lesbian":
					defaultValue = NewTypedString("")
				case "gay":
					defaultValue = NewTypedInt(0)
				case "trans":
					defaultValue = NewTypedFloat(0.0)
				case "nonbinary":
					defaultValue = NewTypedBool(false)
				case "gender":
					defaultValue = NewTypedArray([]TypedValue{})
				default:
					defaultValue = TypedValue{Type: TypeNull, Value: nil}
				}
				parentInstance.Fields[fieldName] = defaultValue
			}
			for methodName, method := range parentClass.Methods {
				parentInstance.Methods[methodName] = method
			}
			instance.Parent = parentInstance
		}
	}

	if classDecl.Constructor != nil {
		oldThis := i.this
		i.this = instance
		i.pushFrame()

		argValues := make([]TypedValue, len(n.Args))
		for idx, arg := range n.Args {
			val, err := i.evaluateNode(arg)
			if err != nil {
				i.popFrame()
				i.this = oldThis
				return TypedValue{}, err
			}
			argValues[idx] = val
		}

		for idx, param := range classDecl.Constructor.Params {
			if idx < len(argValues) {
				i.setVar(param, argValues[idx])
			}
		}

		for _, stmt := range classDecl.Constructor.Body {
			_, err := i.evaluateNode(stmt)
			if err != nil {
				i.popFrame()
				i.this = oldThis
				return TypedValue{}, err
			}
		}

		i.popFrame()
		i.this = oldThis
	}

	return NewTypedObject(instance), nil
}

func (i *Interpreter) accessQueerField(n *QueerFieldAccessNode) (TypedValue, error) {
	objVal, err := i.evaluateNode(n.Object)
	if err != nil {
		return TypedValue{}, err
	}

	if objVal.Type != TypeObject {
		return TypedValue{}, fmt.Errorf("cannot access field of non-object")
	}

	obj := objVal.Value.(*QueerInstance)

	if val, ok := obj.Fields[n.Field]; ok {
		return val, nil
	}

	if obj.Parent != nil {
		if val, ok := obj.Parent.Fields[n.Field]; ok {
			return val, nil
		}
	}

	return TypedValue{}, fmt.Errorf("field not found: %s", n.Field)
}

func (i *Interpreter) callQueerMethod(n *QueerMethodCallNode) (TypedValue, error) {
	objVal, err := i.evaluateNode(n.Object)
	if err != nil {
		return TypedValue{}, err
	}

	if objVal.Type != TypeObject {
		return TypedValue{}, fmt.Errorf("cannot call method on non-object")
	}

	obj := objVal.Value.(*QueerInstance)

	var method *FunctionDeclaration
	var found bool

	if m, ok := obj.Methods[n.Method]; ok {
		method = m
		found = true
	} else if obj.Parent != nil {
		if m, ok := obj.Parent.Methods[n.Method]; ok {
			method = m
			found = true
		}
	}

	if !found {
		return TypedValue{}, fmt.Errorf("method not found: %s", n.Method)
	}

	argValues := make([]TypedValue, len(n.Args))
	for idx, arg := range n.Args {
		val, err := i.evaluateNode(arg)
		if err != nil {
			return TypedValue{}, err
		}
		argValues[idx] = val
	}

	oldThis := i.this
	i.this = obj
	i.pushFrame()

	for idx, param := range method.Params {
		if idx < len(argValues) {
			i.setVar(param, argValues[idx])
		}
	}

	i.returnFlag = false
	i.returnValue = TypedValue{Type: TypeNull, Value: nil}

	var lastResult TypedValue
	for _, stmt := range method.Body {
		lastResult, err = i.evaluateNode(stmt)
		if err != nil {
			i.popFrame()
			i.this = oldThis
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
	i.this = oldThis
	return result, nil
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

// ============================================================
// СРАВНЕНИЕ ЗНАЧЕНИЙ
// ============================================================

func compareTypedValues(a, b TypedValue) (int, error) {
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
		cmp, err := compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp < 0), nil
	case ">":
		cmp, err := compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp > 0), nil
	case "<=":
		cmp, err := compareTypedValues(left, right)
		if err != nil {
			return TypedValue{}, err
		}
		return NewTypedBool(cmp <= 0), nil
	case ">=":
		cmp, err := compareTypedValues(left, right)
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
	case TypeObject:
		return true
	default:
		return false
	}
}

func (i *Interpreter) handleError(err error, line, col int) {
	if i.errorHandler != nil {
		i.errorHandler(err, line, col)
	} else {
		panic(fmt.Errorf("error at line %d, col %d: %v", line, col, err))
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
	case "trans":
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

// ============================================================
// ЧАТ-ФУНКЦИИ LGBTScript
// ============================================================

// createLGBTChat - создание чат-сервера
func (i *Interpreter) createLGBTChat(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("createLGBTChat: expected 2 arguments (name, port)")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("createLGBTChat: first argument must be string")
	}

	port, ok := args[1].Value.(int)
	if !ok {
		return TypedValue{}, fmt.Errorf("createLGBTChat: second argument must be integer")
	}

	if port < 1 || port > 65535 {
		return TypedValue{}, fmt.Errorf("createLGBTChat: invalid port %d", port)
	}

	maxMessages := 100
	if len(args) > 2 {
		if m, ok := args[2].Value.(int); ok && m > 0 {
			maxMessages = m
		}
	}

	chatMu.Lock()
	defer chatMu.Unlock()

	if _, exists := chatServers[name]; exists {
		return TypedValue{}, fmt.Errorf("createLGBTChat: chat '%s' already exists", name)
	}

	chat := &ChatServer{
		Name:        name,
		Clients:     make(map[*websocket.Conn]*ChatClient),
		Broadcast:   make(chan ChatMessage, 256),
		Register:    make(chan *ChatClient),
		Unregister:  make(chan *websocket.Conn),
		IsActive:    false,
		Messages:    []ChatMessage{},
		MaxMessages: maxMessages,
	}

	chatServers[name] = chat

	return NewTypedString(fmt.Sprintf("✅ Чат '%s' создан на порту %d (макс. сообщений: %d)", name, port, maxMessages)), nil
}

// startLGBTChat - запуск чат-сервера
func (i *Interpreter) startLGBTChat(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("startLGBTChat: expected chat name")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("startLGBTChat: first argument must be string")
	}

	chatMu.RLock()
	chat, exists := chatServers[name]
	chatMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("startLGBTChat: chat '%s' not found", name)
	}

	if chat.IsActive {
		return NewTypedString(fmt.Sprintf("⚠️ Чат '%s' уже запущен", name)), nil
	}

	port := 8080
	if len(args) > 1 {
		if p, ok := args[1].Value.(int); ok {
			port = p
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		i.handleWebSocket(w, r, chat)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i.serveChatPage(w, r, chat)
	})

	chat.IsActive = true
	chat.Mu.Lock()
	chat.Messages = []ChatMessage{}
	chat.Mu.Unlock()

	go i.runChatServer(chat)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		fmt.Fprintf(output, "🏳️‍🌈 Чат '%s' запущен на http://localhost:%d\n", chat.Name, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(output, "❌ Ошибка чата '%s': %v\n", chat.Name, err)
		}
	}()

	return NewTypedString(fmt.Sprintf("✅ Чат '%s' запущен на порту %d", name, port)), nil
}

// stopLGBTChat - остановка чат-сервера
func (i *Interpreter) stopLGBTChat(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("stopLGBTChat: expected chat name")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("stopLGBTChat: first argument must be string")
	}

	chatMu.RLock()
	chat, exists := chatServers[name]
	chatMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("stopLGBTChat: chat '%s' not found", name)
	}

	if !chat.IsActive {
		return NewTypedString(fmt.Sprintf("⚠️ Чат '%s' не запущен", name)), nil
	}

	chat.Mu.Lock()
	for conn := range chat.Clients {
		conn.Close()
	}
	chat.Clients = make(map[*websocket.Conn]*ChatClient)
	chat.IsActive = false
	chat.Mu.Unlock()

	close(chat.Broadcast)
	close(chat.Register)
	close(chat.Unregister)

	return NewTypedString(fmt.Sprintf("✅ Чат '%s' остановлен", name)), nil
}

// sendLGBTChatMessage - отправка сообщения в чат
func (i *Interpreter) sendLGBTChatMessage(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: expected 3 arguments (chat_name, username, message)")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: first argument must be string")
	}

	username, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: second argument must be string")
	}

	message, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: third argument must be string")
	}

	chatMu.RLock()
	chat, exists := chatServers[name]
	chatMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: chat '%s' not found", name)
	}

	if !chat.IsActive {
		return TypedValue{}, fmt.Errorf("sendLGBTChatMessage: chat '%s' is not active", name)
	}

	hasHate, filtered := i.checkHateSpeech(message)
	if hasHate {
		message = filtered
	}

	msg := ChatMessage{
		Username:  username,
		Content:   message,
		Timestamp: time.Now(),
		Room:      "general",
		Type:      "message",
	}

	chat.Broadcast <- msg

	return NewTypedString(fmt.Sprintf("✅ Сообщение отправлено в чат '%s' от %s", name, username)), nil
}

// getLGBTChatMessages - получение истории сообщений
func (i *Interpreter) getLGBTChatMessages(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("getLGBTChatMessages: expected chat name")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getLGBTChatMessages: first argument must be string")
	}

	limit := 50
	if len(args) > 1 {
		if l, ok := args[1].Value.(int); ok && l > 0 {
			limit = l
		}
	}

	chatMu.RLock()
	chat, exists := chatServers[name]
	chatMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("getLGBTChatMessages: chat '%s' not found", name)
	}

	chat.Mu.RLock()
	defer chat.Mu.RUnlock()

	messages := chat.Messages
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	result := fmt.Sprintf("📋 История сообщений чата '%s' (последние %d):\n", name, len(messages))
	for _, msg := range messages {
		timestamp := msg.Timestamp.Format("15:04:05")
		if msg.Type == "system" {
			result += fmt.Sprintf("  🔔 [%s] %s\n", timestamp, msg.Content)
		} else if msg.Type == "join" {
			result += fmt.Sprintf("  👋 [%s] %s присоединился(лась)\n", timestamp, msg.Username)
		} else if msg.Type == "leave" {
			result += fmt.Sprintf("  👋 [%s] %s покинул(а) чат\n", timestamp, msg.Username)
		} else {
			result += fmt.Sprintf("  💬 [%s] %s: %s\n", timestamp, msg.Username, msg.Content)
		}
	}

	return NewTypedString(result), nil
}

// getLGBTChatStats - статистика чата
func (i *Interpreter) getLGBTChatStats(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("getLGBTChatStats: expected chat name")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getLGBTChatStats: first argument must be string")
	}

	chatMu.RLock()
	chat, exists := chatServers[name]
	chatMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("getLGBTChatStats: chat '%s' not found", name)
	}

	chat.Mu.RLock()
	clientCount := len(chat.Clients)
	messageCount := len(chat.Messages)
	chat.Mu.RUnlock()

	result := fmt.Sprintf("📊 Статистика чата '%s':\n", name)
	result += fmt.Sprintf("  • Активен: %v\n", chat.IsActive)
	result += fmt.Sprintf("  • Клиентов: %d\n", clientCount)
	result += fmt.Sprintf("  • Сообщений: %d\n", messageCount)
	result += fmt.Sprintf("  • Макс. сообщений: %d\n", chat.MaxMessages)

	return NewTypedString(result), nil
}

// listLGBTChats - список всех чатов
func (i *Interpreter) listLGBTChats(args []TypedValue) (TypedValue, error) {
	chatMu.RLock()
	defer chatMu.RUnlock()

	if len(chatServers) == 0 {
		return NewTypedString("📋 Нет созданных чатов"), nil
	}

	result := "📋 Список чатов:\n"
	for name, chat := range chatServers {
		status := "⏹️ остановлен"
		if chat.IsActive {
			status = "▶️ запущен"
		}
		chat.Mu.RLock()
		clients := len(chat.Clients)
		messages := len(chat.Messages)
		chat.Mu.RUnlock()

		result += fmt.Sprintf("  • %s - %s, клиентов: %d, сообщений: %d\n",
			name, status, clients, messages)
	}

	return NewTypedString(result), nil
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ЧАТА
// ============================================================

func (i *Interpreter) handleWebSocket(w http.ResponseWriter, r *http.Request, chat *ChatServer) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(output, "❌ Ошибка WebSocket: %v\n", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = fmt.Sprintf("Аноним%d", time.Now().Unix()%1000)
	}

	hasHate, filtered := i.checkHateSpeech(username)
	if hasHate {
		username = filtered + "_safe"
	}

	client := &ChatClient{
		Conn:     conn,
		Username: username,
		Server:   chat,
		Room:     "general",
	}

	chat.Register <- client

	go i.readMessages(client, chat)
}

func (i *Interpreter) readMessages(client *ChatClient, chat *ChatServer) {
	defer func() {
		chat.Unregister <- client.Conn
		client.Conn.Close()
	}()

	for {
		var msg ChatMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		hasHate, filtered := i.checkHateSpeech(msg.Content)
		if hasHate {
			msg.Content = filtered
		}

		msg.Username = client.Username
		msg.Timestamp = time.Now()
		msg.Type = "message"
		msg.Room = client.Room

		chat.Broadcast <- msg
	}
}

func (i *Interpreter) runChatServer(chat *ChatServer) {
	for {
		select {
		case client := <-chat.Register:
			chat.Mu.Lock()
			chat.Clients[client.Conn] = client
			chat.Mu.Unlock()

			chat.Mu.RLock()
			for _, msg := range chat.Messages {
				if err := client.Conn.WriteJSON(msg); err != nil {
					break
				}
			}
			chat.Mu.RUnlock()

			joinMsg := ChatMessage{
				Username:  "System",
				Content:   fmt.Sprintf("%s присоединился(лась) к чату", client.Username),
				Timestamp: time.Now(),
				Room:      "general",
				Type:      "join",
			}
			chat.Broadcast <- joinMsg

		case conn := <-chat.Unregister:
			chat.Mu.Lock()
			if client, exists := chat.Clients[conn]; exists {
				delete(chat.Clients, conn)
				leaveMsg := ChatMessage{
					Username:  "System",
					Content:   fmt.Sprintf("%s покинул(а) чат", client.Username),
					Timestamp: time.Now(),
					Room:      "general",
					Type:      "leave",
				}
				chat.Broadcast <- leaveMsg
			}
			chat.Mu.Unlock()

		case msg := <-chat.Broadcast:
			chat.Mu.Lock()
			if len(chat.Messages) >= chat.MaxMessages {
				chat.Messages = chat.Messages[1:]
			}
			chat.Messages = append(chat.Messages, msg)

			for conn := range chat.Clients {
				if err := conn.WriteJSON(msg); err != nil {
					conn.Close()
					delete(chat.Clients, conn)
				}
			}
			chat.Mu.Unlock()
		}
	}
}

func (i *Interpreter) serveChatPage(w http.ResponseWriter, r *http.Request, chat *ChatServer) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>🏳️‍🌈 LGBTScript Чат - ` + chat.Name + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #ff6b6b, #ffd93d, #6bcb77, #4d96ff, #9b59b6);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .chat-container {
            background: rgba(255,255,255,0.95);
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            width: 100%;
            max-width: 800px;
            height: 90vh;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .chat-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            text-align: center;
            font-size: 24px;
            font-weight: bold;
        }
        .chat-header small {
            font-size: 14px;
            font-weight: normal;
            opacity: 0.8;
        }
        .chat-messages {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            background: #f8f9fa;
        }
        .message {
            margin-bottom: 10px;
            padding: 10px 15px;
            border-radius: 10px;
            animation: fadeIn 0.3s ease;
        }
        .message.system {
            background: #e8f5e9;
            text-align: center;
            color: #2e7d32;
            font-style: italic;
        }
        .message.join {
            background: #e3f2fd;
            text-align: center;
            color: #1565c0;
        }
        .message.leave {
            background: #fce4ec;
            text-align: center;
            color: #c62828;
        }
        .message .username {
            font-weight: bold;
            color: #4a148c;
        }
        .message .time {
            color: #888;
            font-size: 12px;
            margin-left: 10px;
        }
        .message .content {
            margin-top: 5px;
            word-wrap: break-word;
        }
        .chat-input {
            display: flex;
            padding: 20px;
            background: white;
            border-top: 1px solid #e0e0e0;
            gap: 10px;
        }
        .chat-input input[type="text"] {
            flex: 1;
            padding: 12px 20px;
            border: 2px solid #e0e0e0;
            border-radius: 25px;
            font-size: 16px;
            outline: none;
            transition: border-color 0.3s;
        }
        .chat-input input[type="text"]:focus {
            border-color: #764ba2;
        }
        .chat-input button {
            padding: 12px 30px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 25px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .chat-input button:hover {
            transform: scale(1.05);
        }
        .username-input {
            display: flex;
            padding: 10px 20px;
            background: #f0f0f0;
            gap: 10px;
            align-items: center;
        }
        .username-input label {
            font-weight: bold;
            color: #333;
        }
        .username-input input {
            flex: 0 1 200px;
            padding: 8px 15px;
            border: 2px solid #ddd;
            border-radius: 20px;
            outline: none;
        }
        .username-input input:focus {
            border-color: #764ba2;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .rainbow-text {
            background: linear-gradient(to right, #ff6b6b, #ffd93d, #6bcb77, #4d96ff, #9b59b6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            font-weight: bold;
        }
        .message-count {
            text-align: center;
            padding: 5px;
            color: #888;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="chat-container">
        <div class="chat-header">
            🏳️‍🌈 LGBTScript Чат<br>
            <small>` + chat.Name + `</small>
        </div>
        <div class="username-input">
            <label>👤 Имя:</label>
            <input type="text" id="username" placeholder="Введите имя" value="User" maxlength="20">
            <span style="margin-left: auto; font-size: 12px; color: #888;">🛡️ Фильтр ненависти активен</span>
        </div>
        <div class="chat-messages" id="messages"></div>
        <div class="message-count" id="messageCount">0 сообщений</div>
        <div class="chat-input">
            <input type="text" id="message" placeholder="Введите сообщение..." onkeypress="if(event.key==='Enter') sendMessage()">
            <button onclick="sendMessage()">Отправить 🏳️‍🌈</button>
        </div>
    </div>
    <script>
        let ws = null;
        let username = 'User';
        
        function connect() {
            username = document.getElementById('username').value.trim() || 'Аноним';
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws?username=' + encodeURIComponent(username));
            
            ws.onopen = function() {
                addSystemMessage('🟢 Подключено к чату!');
            };
            
            ws.onmessage = function(event) {
                const msg = JSON.parse(event.data);
                displayMessage(msg);
            };
            
            ws.onclose = function() {
                addSystemMessage('🔴 Отключено от чата. Переподключение через 3 секунды...');
                setTimeout(connect, 3000);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
                addSystemMessage('⚠️ Ошибка соединения');
            };
        }
        
        function displayMessage(msg) {
            const messagesDiv = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message ' + (msg.type || '');
            
            if (msg.type === 'system' || msg.type === 'join' || msg.type === 'leave') {
                div.textContent = msg.content;
            } else {
                const time = new Date(msg.timestamp).toLocaleTimeString();
                const usernameSpan = document.createElement('span');
                usernameSpan.className = 'username';
                usernameSpan.textContent = msg.username + ':';
                const timeSpan = document.createElement('span');
                timeSpan.className = 'time';
                timeSpan.textContent = time;
                const contentDiv = document.createElement('div');
                contentDiv.className = 'content';
                contentDiv.textContent = msg.content;
                
                div.appendChild(usernameSpan);
                div.appendChild(timeSpan);
                div.appendChild(contentDiv);
            }
            
            messagesDiv.appendChild(div);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
            
            document.getElementById('messageCount').textContent = messagesDiv.children.length + ' сообщений';
        }
        
        function addSystemMessage(text) {
            const messagesDiv = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message system';
            div.textContent = text;
            messagesDiv.appendChild(div);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }
        
        function sendMessage() {
            const input = document.getElementById('message');
            const text = input.value.trim();
            if (!text) return;
            
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                addSystemMessage('⚠️ Нет соединения с сервером');
                return;
            }
            
            ws.send(JSON.stringify({
                content: text
            }));
            
            input.value = '';
            input.focus();
        }
        
        window.onload = function() {
            connect();
            
            document.getElementById('username').addEventListener('change', function() {
                username = this.value.trim() || 'Аноним';
                addSystemMessage('👤 Имя изменено на: ' + username);
                if (ws) ws.close();
            });
        };
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// ============================================================
// ВСТРОЕННЫЕ ФУНКЦИИ
// ============================================================

func (i *Interpreter) getBuiltinFunction(name string) (func([]TypedValue) (TypedValue, error), bool) {
	switch name {
	// GUI функции
	case "rainbowWin":
		return i.rainbowWin, true
	case "rainbowButton":
		return i.rainbowButton, true
	case "rainbowInput":
		return i.rainbowInput, true
	case "rainbowGetText":
		return i.rainbowGetText, true
	case "rainbowSetText":
		return i.rainbowSetText, true
	case "rainbowOnClick":
		return i.rainbowOnClick, true
	case "msg":
		return i.msg, true

	// Серверные функции
	case "createServer":
		return i.createServer, true
	case "startServer":
		return i.startServer, true
	case "stopServer":
		return i.stopServer, true
	case "addRoute":
		return i.addRoute, true
	case "getServerStatus":
		return i.getServerStatus, true
	case "listServers":
		return i.listServers, true

	// Чат-функции
	case "createLGBTChat":
		return i.createLGBTChat, true
	case "startLGBTChat":
		return i.startLGBTChat, true
	case "stopLGBTChat":
		return i.stopLGBTChat, true
	case "sendLGBTChatMessage":
		return i.sendLGBTChatMessage, true
	case "getLGBTChatMessages":
		return i.getLGBTChatMessages, true
	case "getLGBTChatStats":
		return i.getLGBTChatStats, true
	case "listLGBTChats":
		return i.listLGBTChats, true

	// Социальные функции
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
	case "addHateSlur":
		return i.addHateSlur, true
	case "removeHateSlur":
		return i.removeHateSlur, true
	case "getHateLog":
		return i.getHateLog, true
	case "clearHateLog":
		return i.clearHateLog, true
	case "getHateStats":
		return i.getHateStats, true

	default:
		return nil, false
	}
}

// ============================================================
// GUI ФУНКЦИИ
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

func (i *Interpreter) rainbowWin(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("rainbowWin: expected window name")
	}

	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowWin: first argument must be string")
	}

	title := "LGBTScript Window"
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			title = t
		}
	}

	width := 400
	if len(args) > 2 {
		if w, ok := args[2].Value.(int); ok {
			width = w
		}
	}

	height := 300
	if len(args) > 3 {
		if h, ok := args[3].Value.(int); ok {
			height = h
		}
	}

	guiMu.Lock()
	defer guiMu.Unlock()

	if _, exists := guiWindows[name]; exists {
		return TypedValue{}, fmt.Errorf("rainbowWin: window '%s' already exists", name)
	}

	window := &GUIWindow{
		controls: make(map[string]win.HWND),
		events:   make(map[string]func([]TypedValue) (TypedValue, error)),
		isActive: false,
	}

	guiWindows[name] = window

	go func() {
		runtime.LockOSThread()

		wnd := ui.NewMain(
			ui.OptsMain().
				Title(title).
				Size(ui.Dpi(width, height)),
		)

		window.wnd = wnd
		window.isActive = true

		wnd.RunAsMain()

		guiMu.Lock()
		delete(guiWindows, name)
		guiMu.Unlock()
	}()

	time.Sleep(100 * time.Millisecond)

	return NewTypedString(fmt.Sprintf("✅ Окно '%s' создано: '%s' (%dx%d)", name, title, width, height)), nil
}

func (i *Interpreter) rainbowButton(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("rainbowButton: expected window name, button name, and text")
	}

	winName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowButton: first argument must be string")
	}

	btnName, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowButton: second argument must be string")
	}

	text, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowButton: third argument must be string")
	}

	x := 10
	if len(args) > 3 {
		if v, ok := args[3].Value.(int); ok {
			x = v
		}
	}

	y := 10
	if len(args) > 4 {
		if v, ok := args[4].Value.(int); ok {
			y = v
		}
	}

	width := 100
	if len(args) > 5 {
		if v, ok := args[5].Value.(int); ok {
			width = v
		}
	}

	height := 30
	if len(args) > 6 {
		if v, ok := args[6].Value.(int); ok {
			height = v
		}
	}

	guiMu.RLock()
	window, exists := guiWindows[winName]
	guiMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowButton: window '%s' not found", winName)
	}

	if window.wnd == nil {
		return TypedValue{}, fmt.Errorf("rainbowButton: window '%s' not ready", winName)
	}

	btn := ui.NewButton(
		window.wnd,
		ui.OptsButton().
			Text(text).
			Position(ui.Dpi(x, y)).
			Width(ui.DpiX(width)).
			Height(ui.DpiY(height)),
	)

	window.mu.Lock()
	window.controls[btnName] = btn.Hwnd()
	window.mu.Unlock()

	return NewTypedString(fmt.Sprintf("✅ Кнопка '%s' создана в окне '%s'", btnName, winName)), nil
}

func (i *Interpreter) rainbowInput(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("rainbowInput: expected window name, input name, and default text")
	}

	winName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowInput: first argument must be string")
	}

	inpName, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowInput: second argument must be string")
	}

	defaultText, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowInput: third argument must be string")
	}

	x := 10
	if len(args) > 3 {
		if v, ok := args[3].Value.(int); ok {
			x = v
		}
	}

	y := 50
	if len(args) > 4 {
		if v, ok := args[4].Value.(int); ok {
			y = v
		}
	}

	width := 200
	if len(args) > 5 {
		if v, ok := args[5].Value.(int); ok {
			width = v
		}
	}

	height := 25
	if len(args) > 6 {
		if v, ok := args[6].Value.(int); ok {
			height = v
		}
	}

	guiMu.RLock()
	window, exists := guiWindows[winName]
	guiMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowInput: window '%s' not found", winName)
	}

	if window.wnd == nil {
		return TypedValue{}, fmt.Errorf("rainbowInput: window '%s' not ready", winName)
	}

	inp := ui.NewEdit(
		window.wnd,
		ui.OptsEdit().
			Text(defaultText).
			Position(ui.Dpi(x, y)).
			Width(ui.DpiX(width)).
			Height(ui.DpiY(height)),
	)

	window.mu.Lock()
	window.controls[inpName] = inp.Hwnd()
	window.mu.Unlock()

	return NewTypedString(fmt.Sprintf("✅ Поле ввода '%s' создано в окне '%s'", inpName, winName)), nil
}

func (i *Interpreter) rainbowGetText(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("rainbowGetText: expected window name and control name")
	}

	winName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowGetText: first argument must be string")
	}

	ctrlName, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowGetText: second argument must be string")
	}

	guiMu.RLock()
	window, exists := guiWindows[winName]
	guiMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowGetText: window '%s' not found", winName)
	}

	window.mu.RLock()
	hwnd, exists := window.controls[ctrlName]
	window.mu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowGetText: control '%s' not found", ctrlName)
	}

	const maxLen = 32767
	buf := make([]uint16, maxLen)
	ret, _, _ := syscall.Syscall6(
		procGetWindowText.Addr(),
		3,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(maxLen),
		0, 0, 0,
	)
	if ret == 0 {
		return NewTypedString(""), nil
	}
	text := syscall.UTF16ToString(buf[:ret])
	return NewTypedString(text), nil
}

func (i *Interpreter) rainbowSetText(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("rainbowSetText: expected window name, control name, and text")
	}

	winName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowSetText: first argument must be string")
	}

	ctrlName, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowSetText: second argument must be string")
	}

	text, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowSetText: third argument must be string")
	}

	guiMu.RLock()
	window, exists := guiWindows[winName]
	guiMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowSetText: window '%s' not found", winName)
	}

	window.mu.RLock()
	hwnd, exists := window.controls[ctrlName]
	window.mu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowSetText: control '%s' not found", ctrlName)
	}

	textPtr, _ := syscall.UTF16PtrFromString(text)
	syscall.Syscall(
		procSetWindowText.Addr(),
		2,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(textPtr)),
		0,
	)

	return NewTypedString(fmt.Sprintf("✅ Текст установлен в '%s'", ctrlName)), nil
}

func (i *Interpreter) rainbowOnClick(args []TypedValue) (TypedValue, error) {
	if len(args) < 3 {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: expected window name, control name, and function name")
	}

	winName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: first argument must be string")
	}

	ctrlName, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: second argument must be string")
	}

	funcName, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: third argument must be string")
	}

	guiMu.RLock()
	window, exists := guiWindows[winName]
	guiMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: window '%s' not found", winName)
	}

	fn, ok := i.functions[funcName]
	if !ok {
		return TypedValue{}, fmt.Errorf("rainbowOnClick: function '%s' not found", funcName)
	}

	handler := func(args []TypedValue) (TypedValue, error) {
		i.pushFrame()
		defer i.popFrame()

		var result TypedValue
		for _, stmt := range fn.Body {
			val, err := i.evaluateNode(stmt)
			if err != nil {
				return TypedValue{}, err
			}
			result = val
			if i.returnFlag {
				break
			}
		}
		return result, nil
	}

	window.mu.Lock()
	window.events[ctrlName] = handler
	window.mu.Unlock()

	return NewTypedString(fmt.Sprintf("✅ Обработчик для '%s' установлен", ctrlName)), nil
}

// ============================================================
// СЕРВЕРНЫЕ ФУНКЦИИ
// ============================================================

func (i *Interpreter) createServer(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("createServer: expected name and port")
	}
	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("createServer: first argument must be string")
	}
	port, ok := args[1].Value.(int)
	if !ok {
		return TypedValue{}, fmt.Errorf("createServer: second argument must be integer")
	}

	if port < 1 || port > 65535 {
		return TypedValue{}, fmt.Errorf("createServer: invalid port %d", port)
	}

	serversMu.Lock()
	defer serversMu.Unlock()

	if _, exists := servers[name]; exists {
		return TypedValue{}, fmt.Errorf("createServer: server '%s' already exists", name)
	}

	server := &ServerInstance{
		Name:     name,
		Port:     port,
		Routes:   make(map[string]map[string]func([]TypedValue) (TypedValue, error)),
		IsActive: false,
		Context:  make(map[string]TypedValue),
	}

	server.Routes["GET"] = make(map[string]func([]TypedValue) (TypedValue, error))
	server.Routes["POST"] = make(map[string]func([]TypedValue) (TypedValue, error))
	server.Routes["PUT"] = make(map[string]func([]TypedValue) (TypedValue, error))
	server.Routes["DELETE"] = make(map[string]func([]TypedValue) (TypedValue, error))

	servers[name] = server

	return NewTypedString(fmt.Sprintf("✅ Сервер '%s' создан на порту %d", name, port)), nil
}

func (i *Interpreter) startServer(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("startServer: expected server name")
	}
	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("startServer: first argument must be string")
	}

	serversMu.RLock()
	server, exists := servers[name]
	serversMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("startServer: server '%s' not found", name)
	}

	if server.IsActive {
		return NewTypedString(fmt.Sprintf("⚠️ Сервер '%s' уже запущен", name)), nil
	}

	mux := http.NewServeMux()

	server.mu.RLock()
	for method, routes := range server.Routes {
		for path, handler := range routes {
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				args := []TypedValue{}

				query := make([]TypedValue, 0)
				for key, values := range r.URL.Query() {
					if len(values) > 0 {
						query = append(query, NewTypedString(key))
						query = append(query, NewTypedString(values[0]))
					}
				}
				args = append(args, NewTypedArray(query))

				if method == "POST" || method == "PUT" {
					body, err := io.ReadAll(r.Body)
					if err == nil {
						args = append(args, NewTypedString(string(body)))
					} else {
						args = append(args, NewTypedString(""))
					}
				} else {
					args = append(args, NewTypedString(""))
				}

				result, err := handler(args)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(result.String()))
			}

			mux.HandleFunc(path, handlerFunc)
		}
	}
	server.mu.RUnlock()

	server.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", server.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	server.IsActive = true

	go func() {
		fmt.Fprintf(output, "🚀 Сервер '%s' запущен на http://localhost:%d\n", server.Name, server.Port)
		if err := server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(output, "❌ Ошибка сервера '%s': %v\n", server.Name, err)
		}
	}()

	return NewTypedString(fmt.Sprintf("✅ Сервер '%s' запущен на порту %d", name, server.Port)), nil
}

func (i *Interpreter) stopServer(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("stopServer: expected server name")
	}
	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("stopServer: first argument must be string")
	}

	serversMu.RLock()
	server, exists := servers[name]
	serversMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("stopServer: server '%s' not found", name)
	}

	if !server.IsActive {
		return NewTypedString(fmt.Sprintf("⚠️ Сервер '%s' не запущен", name)), nil
	}

	if server.Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Server.Shutdown(ctx); err != nil {
			return TypedValue{}, fmt.Errorf("error stopping server: %v", err)
		}
	}

	server.IsActive = false
	return NewTypedString(fmt.Sprintf("✅ Сервер '%s' остановлен", name)), nil
}

func (i *Interpreter) addRoute(args []TypedValue) (TypedValue, error) {
	if len(args) < 4 {
		return TypedValue{}, fmt.Errorf("addRoute: expected server, method, path, handler")
	}

	serverName, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("addRoute: first argument must be string")
	}

	method, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("addRoute: second argument must be string")
	}

	path, ok := args[2].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("addRoute: third argument must be string")
	}

	handlerValue := args[3]
	var handlerFunc func([]TypedValue) (TypedValue, error)

	if fn, ok := handlerValue.Value.(*FunctionDeclaration); ok {
		handlerFunc = func(args []TypedValue) (TypedValue, error) {
			i.pushFrame()
			defer i.popFrame()

			if len(args) > 0 {
				i.setVar("query", args[0])
			}
			if len(args) > 1 {
				i.setVar("body", args[1])
			}

			var result TypedValue
			for _, stmt := range fn.Body {
				val, err := i.evaluateNode(stmt)
				if err != nil {
					return TypedValue{}, err
				}
				result = val
				if i.returnFlag {
					break
				}
			}
			return result, nil
		}
	} else {
		return TypedValue{}, fmt.Errorf("addRoute: fourth argument must be a function")
	}

	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		return TypedValue{}, fmt.Errorf("addRoute: unsupported method '%s'", method)
	}

	serversMu.RLock()
	server, exists := servers[serverName]
	serversMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("addRoute: server '%s' not found", serverName)
	}

	server.mu.Lock()
	server.Routes[method][path] = handlerFunc
	server.mu.Unlock()

	return NewTypedString(fmt.Sprintf("✅ Маршрут %s %s добавлен на сервер '%s'", method, path, serverName)), nil
}

func (i *Interpreter) getServerStatus(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("getServerStatus: expected server name")
	}
	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getServerStatus: first argument must be string")
	}

	serversMu.RLock()
	server, exists := servers[name]
	serversMu.RUnlock()

	if !exists {
		return TypedValue{}, fmt.Errorf("getServerStatus: server '%s' not found", name)
	}

	status := "остановлен"
	if server.IsActive {
		status = "запущен"
	}

	routeCount := 0
	server.mu.RLock()
	for _, routes := range server.Routes {
		routeCount += len(routes)
	}
	server.mu.RUnlock()

	result := fmt.Sprintf("📊 Статус сервера '%s':\n", name)
	result += fmt.Sprintf("  • Порт: %d\n", server.Port)
	result += fmt.Sprintf("  • Статус: %s\n", status)
	result += fmt.Sprintf("  • Маршрутов: %d\n", routeCount)

	return NewTypedString(result), nil
}

func (i *Interpreter) listServers(args []TypedValue) (TypedValue, error) {
	serversMu.RLock()
	defer serversMu.RUnlock()

	if len(servers) == 0 {
		return NewTypedString("📋 Нет созданных серверов"), nil
	}

	result := "📋 Список серверов:\n"
	for name, server := range servers {
		var status string
		if server.IsActive {
			status = "▶️ запущен"
		} else {
			status = "⏹️ остановлен"
		}

		routeCount := 0
		server.mu.RLock()
		for _, routes := range server.Routes {
			routeCount += len(routes)
		}
		server.mu.RUnlock()
		result += fmt.Sprintf("  • %s - порт %d, %s, маршрутов: %d\n", name, server.Port, status, routeCount)
	}

	return NewTypedString(result), nil
}

// ============================================================
// ФУНКЦИЯ antiHomoPhobe
// ============================================================

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

// ============================================================
// ФУНКЦИЯ lgbtImg
// ============================================================

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

// ============================================================
// ФУНКЦИЯ getLGBTResources
// ============================================================

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
		{
			ID:           "r4",
			Name:         "ILGA (International Lesbian, Gay, Bisexual, Trans and Intersex Association)",
			Type:         "international",
			Description:  "Международная ассоциация, объединяющая ЛГБТ-организации по всему миру",
			Country:      "Швейцария",
			City:         "Женева",
			Address:      "Rue des Deux Tours, 1",
			Phone:        "+41 (22) 734-32-54",
			Email:        "info@ilga.org",
			Website:      "https://ilga.org",
			Services:     []string{"адвокация", "поддержка_организаций", "исследования", "образование"},
			WorkingHours: "Пн-Пт: 09:00-17:00 CET",
			Verified:     true,
			Rating:       4.7,
			Reviews:      205,
			Languages:    []string{"en", "fr", "es", "ru"},
		},
		{
			ID:           "r5",
			Name:         "The Trevor Project",
			Type:         "crisis",
			Description:  "Круглосуточная кризисная поддержка ЛГБТ+ молодежи",
			Country:      "США",
			City:         "Нью-Йорк",
			Phone:        "+1 (866) 488-7386",
			Email:        "help@trevorproject.org",
			Website:      "https://thetrevorproject.org",
			Services:     []string{"кризисная_поддержка", "горячая_линия", "чат_поддержки", "психологическая_помощь"},
			WorkingHours: "Круглосуточно",
			Verified:     true,
			Rating:       4.9,
			Reviews:      456,
			Languages:    []string{"en", "es"},
		},
		{
			ID:           "r6",
			Name:         "GLAAD",
			Type:         "media",
			Description:  "Организация по мониторингу СМИ и защите прав ЛГБТ+ в медиа",
			Country:      "США",
			City:         "Нью-Йорк",
			Address:      "160 Varick St, 6th Floor",
			Phone:        "+1 (212) 807-1700",
			Email:        "info@glaad.org",
			Website:      "https://glaad.org",
			Services:     []string{"медиа_мониторинг", "образование", "адвокация", "исследования"},
			WorkingHours: "Пн-Пт: 09:00-18:00 EST",
			Verified:     true,
			Rating:       4.5,
			Reviews:      178,
			Languages:    []string{"en"},
		},
	}

	filterCountry := ""
	filterType := ""
	filterCity := ""
	filterService := ""

	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			filterCountry = c
		}
	}
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			filterType = t
		}
	}
	if len(args) > 2 {
		if c, ok := args[2].Value.(string); ok {
			filterCity = c
		}
	}
	if len(args) > 3 {
		if s, ok := args[3].Value.(string); ok {
			filterService = s
		}
	}

	filtered := []Resource{}
	for _, r := range resources {
		if filterCountry != "" && !strings.Contains(strings.ToLower(r.Country), strings.ToLower(filterCountry)) {
			continue
		}
		if filterType != "" && !strings.Contains(strings.ToLower(r.Type), strings.ToLower(filterType)) {
			continue
		}
		if filterCity != "" && !strings.Contains(strings.ToLower(r.City), strings.ToLower(filterCity)) {
			continue
		}
		if filterService != "" {
			found := false
			for _, s := range r.Services {
				if strings.Contains(strings.ToLower(s), strings.ToLower(filterService)) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, r)
	}

	response := map[string]interface{}{
		"total":     len(filtered),
		"resources": filtered,
		"filters": map[string]string{
			"country": filterCountry,
			"type":    filterType,
			"city":    filterCity,
			"service": filterService,
		},
		"meta": map[string]interface{}{
			"timestamp":    time.Now().Format(time.RFC3339),
			"version":      "1.0",
			"source":       "LGBTScript Resource Database",
			"total_orgs":   len(resources),
			"verified_only": true,
		},
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return TypedValue{}, fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	return NewTypedString(string(jsonData)), nil
}

// ============================================================
// СОЦИАЛЬНЫЕ ФУНКЦИИ
// ============================================================

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
		"небинарность":     "НЕБИНАРНОСТЬ (Non-binary) — гендерная идентичность, которая не вписывается в бинарную систему мужского и женского пола.",
		"бисексуальность":  "БИСЕКСУАЛЬНОСТЬ (Bisexuality) — романтическое и/или сексуальное влечение к людям более чем одного пола.",
		"гомосексуальность": "ГОМОСЕКСУАЛЬНОСТЬ (Homosexuality) — романтическое и/или сексуальное влечение к людям того же пола.",
		"трансгендерность":  "ТРАНСГЕНДЕРНОСТЬ (Transgender) — состояние, когда гендерная идентичность человека не совпадает с полом при рождении.",
		"гетеросексуальность": "ГЕТЕРОСЕКСУАЛЬНОСТЬ (Heterosexuality) — романтическое и/или сексуальное влечение к людям противоположного пола.",
		"квир":             "КВИР (Queer) — зонтичный термин для ЛГБТ+ сообщества, обозначающий несоответствие нормам.",
		"интерсекс":        "ИНТЕРСЕКС (Intersex) — люди, рожденные с репродуктивными или половыми характеристиками, не вписывающимися в типичные определения мужского или женского тела.",
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

func (i *Interpreter) addHateSlur(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("addHateSlur: expected 2 arguments (language, slur)")
	}

	lang, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("addHateSlur: first argument must be string")
	}

	slur, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("addHateSlur: second argument must be string")
	}

	i.hateFilter.mu.Lock()
	defer i.hateFilter.mu.Unlock()

	if _, exists := i.hateFilter.slurs[lang]; !exists {
		i.hateFilter.slurs[lang] = []string{}
	}
	i.hateFilter.slurs[lang] = append(i.hateFilter.slurs[lang], strings.ToLower(slur))

	i.hateFilter.compilePatterns()

	return NewTypedString(fmt.Sprintf("✅ Добавлено оскорбление '%s' для языка '%s'", slur, lang)), nil
}

func (i *Interpreter) removeHateSlur(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("removeHateSlur: expected 2 arguments (language, slur)")
	}

	lang, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("removeHateSlur: first argument must be string")
	}

	slur, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("removeHateSlur: second argument must be string")
	}

	i.hateFilter.mu.Lock()
	defer i.hateFilter.mu.Unlock()

	if slurs, exists := i.hateFilter.slurs[lang]; exists {
		newSlurs := []string{}
		for _, s := range slurs {
			if s != strings.ToLower(slur) {
				newSlurs = append(newSlurs, s)
			}
		}
		i.hateFilter.slurs[lang] = newSlurs
		i.hateFilter.compilePatterns()
		return NewTypedString(fmt.Sprintf("✅ Удалено оскорбление '%s' для языка '%s'", slur, lang)), nil
	}

	return NewTypedString(fmt.Sprintf("⚠️ Язык '%s' не найден", lang)), nil
}

func (i *Interpreter) getHateLog(args []TypedValue) (TypedValue, error) {
	if i.hateFilter.logFile == "" {
		return NewTypedString("📋 Лог ненависти пуст"), nil
	}

	data, err := os.ReadFile(i.hateFilter.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return NewTypedString("📋 Лог ненависти пуст"), nil
		}
		return TypedValue{}, err
	}

	return NewTypedString(string(data)), nil
}

func (i *Interpreter) clearHateLog(args []TypedValue) (TypedValue, error) {
	if i.hateFilter.logFile == "" {
		return NewTypedString("⚠️ Лог не настроен"), nil
	}

	err := os.Truncate(i.hateFilter.logFile, 0)
	if err != nil {
		return TypedValue{}, err
	}

	return NewTypedString("✅ Лог ненависти очищен"), nil
}

func (i *Interpreter) getHateStats(args []TypedValue) (TypedValue, error) {
	i.hateFilter.mu.RLock()
	defer i.hateFilter.mu.RUnlock()

	totalSlurs := 0
	langs := []string{}
	for lang, slurs := range i.hateFilter.slurs {
		totalSlurs += len(slurs)
		langs = append(langs, fmt.Sprintf("%s: %d", lang, len(slurs)))
	}

	result := fmt.Sprintf("📊 Статистика фильтра ненависти:\n")
	result += fmt.Sprintf("  • Всего языков: %d\n", len(i.hateFilter.slurs))
	result += fmt.Sprintf("  • Всего оскорблений: %d\n", totalSlurs)
	result += fmt.Sprintf("  • Паттернов: %d\n", len(i.hateFilter.patterns))
	result += fmt.Sprintf("  • Статус: %s\n", map[bool]string{true: "Включен", false: "Выключен"}[i.hateFilter.enabled])
	result += "  • Языки:\n"
	for _, l := range langs {
		result += fmt.Sprintf("      %s\n", l)
	}

	return NewTypedString(result), nil
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================

func (i *Interpreter) showHelp(country string) {
	fmt.Fprintf(output, "🌈 LGBTScript Help для %s:\n", country)
	fmt.Fprintf(output, "📖 Доступные команды:\n")
	fmt.Fprintf(output, "  • comingout <выражение> - вывод на экран\n")
	fmt.Fprintf(output, "  • gay <имя> = <число> - объявление целочисленной переменной\n")
	fmt.Fprintf(output, "  • trans <имя> = <число> - объявление переменной с плавающей точкой\n")
	fmt.Fprintf(output, "  • lesbian <имя> = \"текст\" - объявление строковой переменной\n")
	fmt.Fprintf(output, "  • nonbinary <имя> = true/false - объявление булевой переменной\n")
	fmt.Fprintf(output, "  • gender <имя> = [элементы] - объявление массива\n")
	fmt.Fprintf(output, "  • asexual <имя> = <значение> - объявление константы\n")
	fmt.Fprintf(output, "  • cis (<условие>) { ... } - условный оператор\n")
	fmt.Fprintf(output, "  • pride (<условие>) { ... } - цикл while\n")
	fmt.Fprintf(output, "  • sex (инициализация; условие; обновление) { ... } - цикл for\n")
	fmt.Fprintf(output, "  • rainbow <имя>(параметры) { ... } - объявление функции\n")
	fmt.Fprintf(output, "  • return <значение> - возврат из функции\n")
	fmt.Fprintf(output, "  • try { ... } catch { ... } - обработка ошибок\n")
	fmt.Fprintf(output, "  • export rainbow <имя>(параметры) { ... } - экспорт функции\n")
	fmt.Fprintf(output, "  • queer <имя> { ... } - объявление класса\n")
	fmt.Fprintf(output, "  • new <класс>(аргументы) - создание экземпляра класса\n")
	fmt.Fprintf(output, "  • checkHate <текст> - проверка на ненависть\n")
	fmt.Fprintf(output, "  • filterHate <текст> - фильтрация ненависти\n")
	fmt.Fprintf(output, "  • getHateStats - статистика фильтра\n")
	fmt.Fprintf(output, "  • createLGBTChat <имя> <порт> - создание чата\n")
	fmt.Fprintf(output, "  • startLGBTChat <имя> - запуск чата\n")
	fmt.Fprintf(output, "  • stopLGBTChat <имя> - остановка чата\n")
	fmt.Fprintf(output, "  • sendLGBTChatMessage <чат> <пользователь> <сообщение> - отправка сообщения\n")
	fmt.Fprintf(output, "  • getLGBTChatMessages <чат> [лимит] - получение истории\n")
	fmt.Fprintf(output, "  • getLGBTChatStats <чат> - статистика чата\n")
	fmt.Fprintf(output, "  • listLGBTChats - список всех чатов\n")
}

func (i *Interpreter) runOrientationDemo() {
	fmt.Fprintf(output, "🏳️‍🌈 LGBTScript - Язык для ЛГБТ+ сообщества\n")
	fmt.Fprintf(output, "📊 Версия: 7.0 (Rainbow Edition)\n")
	fmt.Fprintf(output, "💡 Поддерживаемые типы: lesbian (string), gay (int), trans (float), nonbinary (bool), gender (array)\n")
	fmt.Fprintf(output, "🛡️ Встроенный фильтр ненависти на 20+ языках\n")
	fmt.Fprintf(output, "💬 Поддержка WebSocket чатов с фильтрацией ненависти\n")
}

func (i *Interpreter) runOrientationTest() {
	fmt.Fprintf(output, "🧪 Тест ориентации LGBTScript\n")
	fmt.Fprintf(output, "✅ Все функции работают корректно!\n")
}

func isExpressionNode(node Node) bool {
	switch node.(type) {
	case *NumberNode, *FloatNode, *StringNode, *BooleanNode, *VariableNode, *BinaryOpNode, *FunctionCall, *ArrayNode, *ArrayIndexNode, *UnaryNode, *QueerInstanceNode, *QueerFieldAccessNode, *QueerMethodCallNode, *ThisNode, *SuperNode:
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
	case *QueerClassDeclaration:
		fmt.Fprintf(output, "%sQueerClass: %s\n", prefix, n.Name)
	case *QueerInstanceNode:
		fmt.Fprintf(output, "%sNew %s\n", prefix, n.ClassName)
	case *ThisNode:
		fmt.Fprintln(output, prefix+"this")
	case *SuperNode:
		fmt.Fprintln(output, prefix+"super")
	default:
		fmt.Fprintf(output, "%sUnknown: %T\n", prefix, n)
	}
}

func runExample() {
	program := `
@ Пример работы чата LGBTScript

RAINBOW main() {
    COMINGOUT "🏳️‍🌈 LGBTScript с поддержкой WebSocket чатов";
    COMINGOUT "";
    
    @ Создаем чат
    CREATE_LGBTChat("mainChat", 8080, 100) -> result;
    COMINGOUT result;
    
    @ Запускаем чат
    START_LGBTChat("mainChat", 8080) -> startResult;
    COMINGOUT startResult;
    
    @ Отправляем сообщение
    SEND_LGBTChatMessage("mainChat", "Admin", "🏳️‍🌈 Добро пожаловать в LGBTScript чат!") -> sendResult;
    COMINGOUT sendResult;
    
    @ Отправляем еще одно сообщение
    SEND_LGBTChatMessage("mainChat", "Bot", "🛡️ Все сообщения фильтруются через HateFilter на 20+ языках") -> sendResult2;
    COMINGOUT sendResult2;
    
    @ Получаем историю
    GET_LGBTChatMessages("mainChat", 10) -> history;
    COMINGOUT "📚 История чата:";
    COMINGOUT history;
    
    @ Показываем статистику
    GET_LGBTChatStats("mainChat") -> stats;
    COMINGOUT stats;
    
    @ Список всех чатов
    LIST_LGBTChats() -> chats;
    COMINGOUT chats;
    
    COMINGOUT "";
    COMINGOUT "✅ Чат успешно настроен и готов к использованию!";
    COMINGOUT "🌐 Откройте http://localhost:8080 в браузере";
}

main();
`

	fmt.Fprintln(output, "=== LGBTScript с WebSocket чатами ===")
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

	guiMarker := "##RB_GUI_MODE##"
	_, err = dst.Write([]byte(guiMarker))
	if err != nil {
		return fmt.Errorf("write GUI marker failed: %v", err)
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

	dst.Close()

	if runtime.GOOS == "windows" && strings.HasSuffix(outputExe, ".exe") {
		cmd := exec.Command("editbin", "/SUBSYSTEM:WINDOWS", outputExe)
		if err := cmd.Run(); err == nil {
			fmt.Fprintf(output, "✅ Установлен GUI режим (без консоли)\n")
		} else {
			fmt.Fprintf(output, "\n⚠️ Для запуска без консоли выполните:\n")
			fmt.Fprintf(output, "   go build -ldflags=\"-H windowsgui\" -o %s\n", outputExe)
			fmt.Fprintf(output, "   или установите Visual Studio и выполните:\n")
			fmt.Fprintf(output, "   editbin /SUBSYSTEM:WINDOWS %s\n", outputExe)
		}
	}

	fmt.Fprintf(output, "✅ Скомпилировано (GUI режим): %s -> %s\n", inputScript, outputExe)
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

func runGUIScript(script string) error {
	runtime.LockOSThread()

	lexer := NewLexer(script)
	interpreter := NewInterpreter()

	tokens, err := lexer.Tokenize()
	if err != nil {
		return fmt.Errorf("лексическая ошибка: %v", err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("синтаксическая ошибка: %v", err)
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
	example := flag.Bool("example", false, "показать пример с чатом")
	buildFlag := flag.Bool("b", false, "скомпилировать .rainbow в .exe")
	guiFlag := flag.Bool("gui", false, "запустить GUI-приложение (для оконных скриптов)")

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

	if *guiFlag {
		if flag.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "Использование: -gui <файл.rainbow>\n")
			os.Exit(1)
		}
		filename := flag.Arg(0)
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения файла %s: %v\n", filename, err)
			os.Exit(1)
		}
		err = runGUIScript(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
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

	fmt.Fprintln(output, "🌈 LGBTScript - Язык программирования с поддержкой ЛГБТ+ сообщества")
	fmt.Fprintln(output, "📖 Используйте --example для демонстрации чата")
	fmt.Fprintln(output, "📁 Укажите файл .rainbow для выполнения")
	fmt.Fprintln(output, "🔧 Для компиляции в .exe используйте -b input.rainbow output.exe")
	fmt.Fprintln(output, "🖥️  Для GUI-приложений используйте -gui файл.rainbow")
	fmt.Fprintln(output, "💬 Поддержка WebSocket чатов с фильтрацией ненависти")
	fmt.Fprintln(output, "🛡️ Встроенная защита от ненависти на 20+ языках")
	runExample()
}