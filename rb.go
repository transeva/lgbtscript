package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
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
	"context"
)

// ---------- Глобальный вывод (консоль + файл) ----------
var output io.Writer = os.Stdout
var isFileExecution = false
var currentFilePath = ""
var debugMode = false

// ---------- Настройка консоли для Windows ----------
func setupConsole() {
	if os.Getenv("OS") == "Windows_NT" {
		fmt.Fprint(output, "\xef\xbb\xbf")
	}
}

// ---------- Токены ----------
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

// ---------- Лексер ----------
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
			"try": true, "catch": true,
			"export": true,
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

		// Директива #inclusive
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

		// Поддержка отрицательных чисел
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

// ---------- AST узлы ----------
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

type Program struct {
	BaseNode
	Statements []Node
}

// ---------- Парсер ----------
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

func (p *Parser) Parse() (*Program, error) {
	program := &Program{BaseNode: BaseNode{Line: 1, Col: 1}}
	for p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program, nil
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
		case "try":
			return p.parseTryCatchStatement()
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

func (p *Parser) parseExportStatement() (Node, error) {
	token := p.peek()
	p.next()
	token2 := p.peek()
	if token2.Type == TOKEN_KEYWORD && token2.Value == "rainbow" {
		fn, err := p.parseFunctionDeclaration()
		if err != nil {
			return nil, err
		}
		if fnDecl, ok := fn.(*FunctionDeclaration); ok {
			fnDecl.Exported = true
			fnDecl.Line = token.Line
			fnDecl.Col = token.Col
		}
		return fn, nil
	}
	return nil, fmt.Errorf("expected function declaration after export")
}

func (p *Parser) parseTryCatchStatement() (Node, error) {
	token := p.peek()
	p.next()
	_, err := p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	tryBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	if p.peek().Type != TOKEN_KEYWORD || p.peek().Value != "catch" {
		return nil, fmt.Errorf("expected 'catch' after try block")
	}
	p.next()
	_, err = p.expect(TOKEN_OPERATOR, "{")
	if err != nil {
		return nil, err
	}

	catchBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &TryCatchStatement{
		BaseNode:   BaseNode{Line: token.Line, Col: token.Col},
		TryBlock:   tryBlock,
		CatchBlock: catchBlock,
	}, nil
}

func (p *Parser) parseArrayAssignment(name string) (Node, error) {
	token := p.peek()
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
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name,
		Index:    index,
		Value:    value,
	}, nil
}

func (p *Parser) parseReturnStatement() (Node, error) {
	token := p.peek()
	p.next()
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &ReturnStatement{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: value}, nil
}

func (p *Parser) parseFunctionDeclaration() (Node, error) {
	token := p.peek()
	p.next()
	nameToken, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}
	name := nameToken.Value

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

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &FunctionDeclaration{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name,
		Params:   params,
		Body:     body,
	}, nil
}

func (p *Parser) parseFunctionCall(name string) (Node, error) {
	token := p.peek()
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

	return &FunctionCall{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Name: name, Args: args}, nil
}

func (p *Parser) parseTypedDeclaration() (Node, error) {
	token := p.peek()
	p.next()
	name, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}

	if p.peek().Value == "=" {
		p.next()
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TOKEN_OPERATOR, ";")
		if err != nil {
			return nil, err
		}
		return &TypedDeclaration{
			BaseNode: BaseNode{Line: token.Line, Col: token.Col},
			Type:     token.Value,
			Name:     name.Value,
			Value:    value,
		}, nil
	}

	_, err = p.expect(TOKEN_OPERATOR, ";")
	if err != nil {
		return nil, err
	}

	var defaultValue Node
	switch token.Value {
	case "lesbian":
		defaultValue = &StringNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: ""}
	case "gay":
		defaultValue = &NumberNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0}
	case "queer":
		defaultValue = &FloatNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0.0}
	case "nonbinary":
		defaultValue = &BooleanNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: false}
	case "gender":
		defaultValue = &ArrayNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Elements: []Node{}}
	default:
		defaultValue = &NumberNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0}
	}

	return &TypedDeclaration{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Type:     token.Value,
		Name:     name.Value,
		Value:    defaultValue,
	}, nil
}

func (p *Parser) parseSexStatement() (Node, error) {
	token := p.peek()
	p.next()

	if p.peek().Value != "(" {
		return nil, fmt.Errorf("expected '(' after SEX, got '%s' at line %d", p.peek().Value, p.peek().Line)
	}
	p.next()

	init, err := p.parseSexPart()
	if err != nil {
		return nil, err
	}
	if p.peek().Value != ";" {
		return nil, fmt.Errorf("expected ';' after initialization in SEX loop at line %d, got '%s'",
			p.peek().Line, p.peek().Value)
	}
	p.next()

	condition, err := p.parseSexPart()
	if err != nil {
		return nil, err
	}
	if condition == nil {
		condition = &BooleanNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: true}
	}
	if p.peek().Value != ";" {
		return nil, fmt.Errorf("expected ';' after condition in SEX loop at line %d, got '%s'",
			p.peek().Line, p.peek().Value)
	}
	p.next()

	update, err := p.parseSexPart()
	if err != nil {
		return nil, err
	}
	if p.peek().Value != ")" {
		return nil, fmt.Errorf("expected ')' after update in SEX loop at line %d, got '%s'",
			p.peek().Line, p.peek().Value)
	}
	p.next()

	if p.peek().Value != "{" {
		return nil, fmt.Errorf("expected '{' after SEX loop header at line %d, got '%s'",
			p.peek().Line, p.peek().Value)
	}
	p.next()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.peek().Value != "}" {
		return nil, fmt.Errorf("expected '}' after SEX loop body at line %d, got '%s'",
			p.peek().Line, p.peek().Value)
	}
	p.next()

	return &SexStatement{
		BaseNode:  BaseNode{Line: token.Line, Col: token.Col},
		Init:      init,
		Condition: condition,
		Update:    update,
		Body:      body,
	}, nil
}

func (p *Parser) parseSexPart() (Node, error) {
	tok := p.peek()
	if tok.Type == TOKEN_EOF || tok.Value == ";" || tok.Value == ")" {
		return nil, nil
	}

	if tok.Type == TOKEN_KEYWORD {
		keyword := tok.Value
		switch keyword {
		case "gay", "lesbian", "queer", "nonbinary", "gender":
			decl, err := p.parseTypedDeclarationNoSemicolon()
			if err != nil {
				return nil, err
			}
			return decl, nil
		}
	}

	return p.parseExpression()
}

func (p *Parser) parseTypedDeclarationNoSemicolon() (Node, error) {
	token := p.peek()
	if token.Type != TOKEN_KEYWORD {
		return nil, fmt.Errorf("expected type keyword, got '%s' at line %d", token.Value, token.Line)
	}
	keyword := token.Value
	if keyword != "gay" && keyword != "lesbian" && keyword != "queer" &&
		keyword != "nonbinary" && keyword != "gender" {
		return nil, fmt.Errorf("expected type keyword (gay, lesbian, queer, nonbinary, gender), got '%s' at line %d",
			keyword, token.Line)
	}
	p.next()

	nameToken := p.peek()
	if nameToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after type, got '%s' at line %d",
			nameToken.Value, nameToken.Line)
	}
	name := nameToken.Value
	p.next()

	var value Node
	var err error
	if p.peek().Value == "=" {
		p.next()
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	} else {
		switch keyword {
		case "gay":
			value = &NumberNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0}
		case "lesbian":
			value = &StringNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: ""}
		case "queer":
			value = &FloatNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0.0}
		case "nonbinary":
			value = &BooleanNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: false}
		case "gender":
			value = &ArrayNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Elements: []Node{}}
		default:
			value = &NumberNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0}
		}
	}

	return &TypedDeclaration{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Type:     keyword,
		Name:     name,
		Value:    value,
	}, nil
}

func (p *Parser) parseIncrementDecrement(name string, postfix bool) (Node, error) {
	token := p.peek()
	op := p.peekNext().Value
	p.next()
	p.next()
	if p.peek().Value == ";" {
		p.next()
	}
	return &IncrementDecrementStatement{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name,
		Operator: op,
		Postfix:  postfix,
	}, nil
}

func (p *Parser) parseAssignment(name string) (Node, error) {
	token := p.peek()
	op := p.peekNext().Value
	p.next()
	p.next()
	
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	if op == "=" {
		if p.peek().Value == ";" {
			p.next()
		}
		return &AssignmentStatement{
			BaseNode: BaseNode{Line: token.Line, Col: token.Col},
			Name:     name,
			Value:    value,
		}, nil
	}
	
	if p.peek().Value == ";" {
		p.next()
	}
	return &AugmentedAssignmentStatement{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name,
		Op:       op[:len(op)-1],
		Value:    value,
	}, nil
}

func (p *Parser) parsePrintStatement() (Node, error) {
	token := p.peek()
	p.next()
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &PrintStatement{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: value}, nil
}

func (p *Parser) parseIfStatement() (Node, error) {
	token := p.peek()
	p.next()
	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpression()
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

	thenBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	var elseIfBlocks []ElseIfBlock
	var elseBlock []Node

	for p.peek().Type == TOKEN_KEYWORD && p.peek().Value == "nocis" {
		p.next()

		var cond Node
		var err error
		if p.peek().Value == "(" {
			_, err = p.expect(TOKEN_OPERATOR, "(")
			if err != nil {
				return nil, err
			}
			cond, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(TOKEN_OPERATOR, ")")
			if err != nil {
				return nil, err
			}
		}

		_, err = p.expect(TOKEN_OPERATOR, "{")
		if err != nil {
			return nil, err
		}

		block, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TOKEN_OPERATOR, "}")
		if err != nil {
			return nil, err
		}

		if cond != nil {
			elseIfBlocks = append(elseIfBlocks, ElseIfBlock{Condition: cond, Block: block})
		} else {
			elseBlock = block
			break
		}
	}

	return &IfStatement{
		BaseNode:     BaseNode{Line: token.Line, Col: token.Col},
		Condition:    condition,
		ThenBlock:    thenBlock,
		ElseIfBlocks: elseIfBlocks,
		ElseBlock:    elseBlock,
	}, nil
}

func (p *Parser) parseWhileStatement() (Node, error) {
	token := p.peek()
	p.next()
	_, err := p.expect(TOKEN_OPERATOR, "(")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpression()
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

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TOKEN_OPERATOR, "}")
	if err != nil {
		return nil, err
	}

	return &WhileStatement{
		BaseNode:  BaseNode{Line: token.Line, Col: token.Col},
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *Parser) parseHelpStatement() (Node, error) {
	token := p.peek()
	p.next()
	country, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &HelpStatement{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Country: country}, nil
}

func (p *Parser) parseOrientationStatement() (Node, error) {
	token := p.peek()
	p.next()
	if p.peek().Value == ";" {
		p.next()
	}
	return &OrientationStatement{BaseNode: BaseNode{Line: token.Line, Col: token.Col}}, nil
}

func (p *Parser) parseBlock() ([]Node, error) {
	var statements []Node
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return statements, nil
}

func (p *Parser) parseConstantDeclaration() (Node, error) {
	token := p.peek()
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
	_, err = p.expect(TOKEN_OPERATOR, ";")
	if err != nil {
		return nil, err
	}
	return &ConstantDeclaration{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name.Value,
		Value:    value,
	}, nil
}

// ---------- Выражения ----------
func (p *Parser) parseExpression() (Node, error) {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() (Node, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "||" {
		op := p.next().Value
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseLogicalAnd() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "&&" {
		op := p.next().Value
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "==" || p.peek().Value == "!=" ||
		p.peek().Value == "<" || p.peek().Value == ">" ||
		p.peek().Value == "<=" || p.peek().Value == ">=" {
		op := p.next().Value
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAdditive() (Node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "+" || p.peek().Value == "-" {
		op := p.next().Value
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMultiplicative() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "*" || p.peek().Value == "/" || p.peek().Value == "%" {
		op := p.next().Value
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.peek().Value == "-" {
		op := p.next().Value
		if p.peek().Type == TOKEN_NUMBER || p.peek().Type == TOKEN_FLOAT {
			num, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			switch n := num.(type) {
			case *NumberNode:
				n.Value = -n.Value
				return n, nil
			case *FloatNode:
				n.Value = -n.Value
				return n, nil
			default:
				return &UnaryNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Op: op, Expr: num}, nil
			}
		}
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0}, Op: op, Expr: expr}, nil
	}
	
	if p.peek().Value == "++" || p.peek().Value == "--" {
		op := p.next().Value
		if p.peek().Type != TOKEN_IDENTIFIER {
			return nil, fmt.Errorf("expected identifier after %s", op)
		}
		name := p.peek().Value
		p.next()
		return &IncrementDecrementStatement{
			BaseNode: BaseNode{Line: p.getCurrentLine(), Col: 0},
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
		p.next()
		val, _ := strconv.Atoi(token.Value)
		return &NumberNode{BaseNode: BaseNode{Line: line, Col: col}, Value: val}, nil
	case TOKEN_FLOAT:
		p.next()
		val, _ := strconv.ParseFloat(token.Value, 64)
		return &FloatNode{BaseNode: BaseNode{Line: line, Col: col}, Value: val}, nil
	case TOKEN_STRING:
		p.next()
		return &StringNode{BaseNode: BaseNode{Line: line, Col: col}, Value: token.Value}, nil
	case TOKEN_ARRAY_OPEN:
		p.next()
		var elements []Node
		for p.peek().Type != TOKEN_ARRAY_CLOSE && p.peek().Type != TOKEN_EOF {
			element, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, element)
			if p.peek().Value == "," {
				p.next()
			}
		}
		_, err := p.expect(TOKEN_ARRAY_CLOSE, "]")
		if err != nil {
			return nil, err
		}
		return &ArrayNode{BaseNode: BaseNode{Line: line, Col: col}, Elements: elements}, nil
	case TOKEN_IDENTIFIER:
		if p.peekNext().Value == "++" || p.peekNext().Value == "--" {
			name := token.Value
			op := p.peekNext().Value
			p.next()
			p.next()
			return &IncrementDecrementStatement{
				BaseNode: BaseNode{Line: line, Col: col},
				Name:     name,
				Operator: op,
				Postfix:  true,
			}, nil
		}
		
		if p.peekNext().Value == "[" {
			p.next()
			name := token.Value
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
			return &ArrayIndexNode{BaseNode: BaseNode{Line: line, Col: col}, Name: name, Index: index}, nil
		}
		if p.peekNext().Value == "(" {
			return p.parseFunctionCall(token.Value)
		}
		p.next()
		return &VariableNode{BaseNode: BaseNode{Line: line, Col: col}, Name: token.Value}, nil
	case TOKEN_KEYWORD:
		keyword := token.Value
		if keyword == "true" {
			p.next()
			return &BooleanNode{BaseNode: BaseNode{Line: line, Col: col}, Value: true}, nil
		} else if keyword == "false" {
			p.next()
			return &BooleanNode{BaseNode: BaseNode{Line: line, Col: col}, Value: false}, nil
		}
		return nil, fmt.Errorf("unexpected keyword: %s", token.Value)
	case TOKEN_OPERATOR:
		if token.Value == "(" {
			p.next()
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(TOKEN_OPERATOR, ")")
			if err != nil {
				return nil, err
			}
			return expr, nil
		}
	}
	return nil, fmt.Errorf("unexpected token: %s", token.Value)
}

// ---------- Система типов ----------
type ValueType int

const (
	TypeString ValueType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeArray
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

func (tv TypedValue) ToString() string {
	return tv.String()
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

// ---------- Структуры для серверов ----------
type ServerInstance struct {
	mu       sync.RWMutex
	Name     string
	Port     int
	Routes   map[string]map[string]func([]TypedValue) (TypedValue, error)
	Server   *http.Server
	IsActive bool
	Context  map[string]TypedValue
}

// Глобальное хранилище серверов
var servers = make(map[string]*ServerInstance)
var serversMu sync.RWMutex

// ============================================
// STABLE DIFFUSION КОНФИГУРАЦИЯ
// ============================================

type SDConfig struct {
	APIURL      string
	APIKey      string
	Model       string
	Timeout     time.Duration
	MaxAttempts int
}

var sdConfig = SDConfig{
	APIURL:      "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image",
	APIKey:      "",
	Model:       "stable-diffusion-xl-1024-v1-0",
	Timeout:     60 * time.Second,
	MaxAttempts: 3,
}

// ---------- Интерпретатор ----------
type callFrame struct {
    vars      map[string]TypedValue
    types     map[string]string
    constants map[string]bool
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
    rand           *rand.Rand
}

func (i *Interpreter) checkForOffensiveTerms(code string) (bool, []string) {
    offensiveTerms := []string{"пидор", "педик", "faggot", "Pouf", "лесбуха", "Bender", "гомосексуалист"}
    foundTerms := []string{}
    
    for _, term := range offensiveTerms {
        if strings.Contains(strings.ToLower(code), strings.ToLower(term)) {
            foundTerms = append(foundTerms, term)
        }
    }
    
    if len(foundTerms) > 0 {
        fmt.Printf("⚠️ Обнаружены потенциально оскорбительные термины: %v\n", foundTerms)
        fmt.Println("💡 Рекомендуем использовать корректную терминологию: ЛГБТ+, гомосексуальный, бисексуальный и т.д.")
        return true, foundTerms
    }
    return false, foundTerms
}

func (i *Interpreter) showSupportMessage() {
    fmt.Println("🏳️‍🌈 Поддержка ЛГБТ+ сообщества:")
    fmt.Println("📞 Горячая линия поддержки: 8-800-XXX-XX-XX")
    fmt.Println("🌐 Ресурсы: https://lgbt-support.org")
    fmt.Println("💙 Вы не одиноки!")
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
        rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
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
		vars:     make(map[string]TypedValue),
		types:    make(map[string]string),
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

// ---------- Sandbox ----------
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

// ============================================
// СЕРВЕРНЫЕ ФУНКЦИИ
// ============================================

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
				val, err := i.Evaluate(stmt)
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
		status := "⏹️ остановлен"
		if server.IsActive {
			status = "▶️ запущен"
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

// ============================================
// ФУНКЦИЯ antiHomoPhobe
// ============================================

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

var (
	user32DLL = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent = user32DLL.NewProc("keybd_event")
)

func keybdEvent(bVk byte, bScan byte, dwFlags uintptr, dwExtraInfo uintptr) {
	procKeybdEvent.Call(uintptr(bVk), uintptr(bScan), dwFlags, dwExtraInfo)
}

// ============================================
// ФУНКЦИИ ДЛЯ ГЕНЕРАЦИИ ИЗОБРАЖЕНИЙ (STABLE DIFFUSION)
// ============================================

func (i *Interpreter) setSDKey(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("setSDKey: expected API key")
	}
	key, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("setSDKey: first argument must be string")
	}
	sdConfig.APIKey = key
	return NewTypedString("✅ API ключ Stable Diffusion установлен"), nil
}

func (i *Interpreter) setSDModel(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("setSDModel: expected model name")
	}
	model, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("setSDModel: first argument must be string")
	}
	sdConfig.Model = model
	return NewTypedString(fmt.Sprintf("✅ Модель изменена на: %s", model)), nil
}

func (i *Interpreter) generateLGBTImage(args []TypedValue) (TypedValue, error) {
	filename := "lgbt.png"
	if len(args) > 0 {
		if name, ok := args[0].Value.(string); ok && name != "" {
			if !strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".jpeg") {
				name = name + ".png"
			}
			filename = name
		}
	}
	
	lgbtPrompts := []string{
		"beautiful rainbow flag with red orange yellow green blue purple stripes waving in wind, vibrant colors, realistic, high quality photography, 8k, detailed",
		"colorful rainbow flag flying in blue sky, bright colors, realistic photography, high quality, peaceful, beautiful",
		"rainbow flag with six horizontal stripes, waving gently, vibrant colors, realistic, high quality, professional photography",
		"rainbow pride flag flowing in wind, bright vivid colors, realistic, high quality, 4k, detailed",
		"flag with light blue pink and white horizontal stripes waving in wind, pastel colors, realistic, high quality photography, soft, peaceful",
		"transgender flag with blue pink and white stripes, waving gracefully, pastel colors, realistic, high quality, beautiful",
		"light blue pink white flag waving, soft pastel colors, realistic photography, high quality, serene",
		"flag with pink purple blue horizontal stripes waving in wind, bright colors, realistic, high quality photography, vibrant",
		"bisexual flag with pink purple and blue stripes, waving, bright colors, realistic, high quality, beautiful",
		"pink purple blue flag waving in wind, vibrant colors, realistic photography, high quality",
		"flag with pink yellow blue horizontal stripes waving in wind, bright colors, realistic, high quality photography, colorful",
		"flag with yellow white purple black horizontal stripes waving in wind, modern design, realistic, high quality photography",
		"flag with black grey white purple horizontal stripes waving in wind, minimalist design, realistic, high quality photography",
		"flag with orange white pink horizontal stripes waving in wind, bright colors, realistic, high quality photography",
		"colorful flag with rainbow stripes and triangle, waving, bright colors, realistic, high quality photography, modern",
		"abstract art with rainbow colors, flowing, beautiful, high quality, artistic, colorful, 8k",
		"heart made of rainbow colors, vibrant, beautiful, realistic, high quality, 4k",
		"rainbow colored waves, flowing together, harmonious, beautiful, artistic, high quality",
		"colorful celebration, confetti, rainbow colors, joyful, vibrant, realistic, high quality",
		"rainbow path through a field, colorful, beautiful, realistic, high quality photography",
		"cityscape with rainbow lights, colorful, modern, realistic, high quality, 8k",
		"rainbow colors merging together, beautiful, artistic, high quality, detailed",
	}
	
	promptIndex := i.rand.Intn(len(lgbtPrompts))
	prompt := lgbtPrompts[promptIndex]
	
	validSizes := [][2]int{
		{1024, 1024}, {1152, 896}, {1216, 832}, {1344, 768},
		{1536, 640}, {640, 1536}, {768, 1344}, {832, 1216}, {896, 1152},
	}
	
	sizeIndex := i.rand.Intn(len(validSizes))
	width, height := validSizes[sizeIndex][0], validSizes[sizeIndex][1]
	
	steps := 25 + i.rand.Intn(16)
	cfgScale := 6.0 + float64(i.rand.Intn(31))/10.0
	
	qualityBoost := []string{"", "masterpiece", "best quality", "ultra detailed", "photorealistic"}
	quality := qualityBoost[i.rand.Intn(len(qualityBoost))]
	if quality != "" {
		prompt = prompt + ", " + quality
	}
	
	if sdConfig.APIKey == "" {
		return i.generateLocalLGBTImageWithName(prompt, width, height, filename)
	}
	
	return i.generateWithStabilityAIWithName(prompt, width, height, steps, cfgScale, filename)
}

func (i *Interpreter) generateWithStabilityAIWithName(prompt string, width, height, steps int, cfgScale float64, filename string) (TypedValue, error) {
	if sdConfig.APIKey == "" {
		return TypedValue{}, fmt.Errorf("Stability AI API key not set. Use setSDKey() first")
	}
	
	validSizes := [][2]int{
		{1024, 1024}, {1152, 896}, {1216, 832}, {1344, 768}, 
		{1536, 640}, {640, 1536}, {768, 1344}, {832, 1216}, {896, 1152},
	}
	
	isValid := false
	for _, size := range validSizes {
		if width == size[0] && height == size[1] {
			isValid = true
			break
		}
	}
	
	if !isValid {
		bestSize := validSizes[0]
		bestDiff := 999999999
		
		for _, size := range validSizes {
			diff := (width-size[0])*(width-size[0]) + (height-size[1])*(height-size[1])
			if diff < bestDiff {
				bestDiff = diff
				bestSize = size
			}
		}
		
		width, height = bestSize[0], bestSize[1]
		fmt.Fprintf(output, "⚠️ Размеры скорректированы до %dx%d (допустимые для SDXL)\n", width, height)
	}
	
	requestBody := map[string]interface{}{
		"text_prompts": []map[string]interface{}{
			{
				"text":   prompt,
				"weight": 1.0,
			},
		},
		"cfg_scale":    cfgScale,
		"height":       height,
		"width":        width,
		"samples":      1,
		"steps":        steps,
		"style_preset": "photographic",
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error marshaling request: %v", err)
	}
	
	req, err := http.NewRequest("POST", sdConfig.APIURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return TypedValue{}, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sdConfig.APIKey))
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{Timeout: sdConfig.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return TypedValue{}, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TypedValue{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TypedValue{}, fmt.Errorf("error decoding response: %v", err)
	}
	
	artifacts, ok := result["artifacts"].([]interface{})
	if !ok || len(artifacts) == 0 {
		return TypedValue{}, fmt.Errorf("no image generated")
	}
	
	artifact := artifacts[0].(map[string]interface{})
	base64Image, ok := artifact["base64"].(string)
	if !ok || base64Image == "" {
		return TypedValue{}, fmt.Errorf("no base64 image in response")
	}
	
	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error decoding image: %v", err)
	}
	
	if err := os.WriteFile(filename, imageData, 0644); err != nil {
		return TypedValue{}, fmt.Errorf("error saving image: %v", err)
	}
	
	resultMsg := fmt.Sprintf("✅ Изображение сгенерировано и сохранено как '%s'\n", filename)
	resultMsg += fmt.Sprintf("📝 Промпт: %s\n", prompt)
	resultMsg += fmt.Sprintf("📐 Размер: %dx%d\n", width, height)
	resultMsg += fmt.Sprintf("🔄 Шагов: %d\n", steps)
	resultMsg += "🏳️‍🌈 ЛГБТ-тематика добавлена автоматически"
	
	return NewTypedString(resultMsg), nil
}

func (i *Interpreter) generateLocalLGBTImageWithName(prompt string, width, height int, filename string) (TypedValue, error) {
	asciiArts := []string{
		`   🌈🌈🌈🌈🌈🌈🌈
   🌈🌈🌈🌈🌈🌈🌈
  🌈🌈🌈🌈🌈🌈🌈
 🌈🌈🌈🌈🌈🌈🌈
🌈🌈🌈🌈🌈🌈🌈
   🏳️‍🌈 LOVE IS LOVE 🏳️‍🌈
   💙💚💛🧡❤️💜`,
		`   ❤️🧡💛💚💙💜
   ❤️🧡💛💚💙💜
  ❤️🧡💛💚💙💜
 ❤️🧡💛💚💙💜
❤️🧡💛💚💙💜
   💜 LOVE WINS 💜
   🏳️‍🌈🏳️‍🌈🏳️‍🌈`,
		`   🔴🟠🟡🟢🔵🟣
   🔴🟠🟡🟢🔵🟣
   🔴🟠🟡🟢🔵🟣
  🔴🟠🟡🟢🔵🟣
 🔴🟠🟡🟢🔵🟣
🔴🟠🟡🟢🔵🟣
   🏳️‍🌈 PRIDE 🏳️‍🌈`,
		`       ✨
     ✨✨✨
   ✨✨✨✨✨
 ✨✨✨✨✨✨✨
   🌈🌈🌈🌈🌈
   🏳️‍🌈🌈🏳️‍🌈
   💖🌈🦄🌈💖`,
		`   🌸🌺🌻🌹🌷🌼
   🌸🌺🌻🌹🌷🌼
  🌸🌺🌻🌹🌷🌼
 🌸🌺🌻🌹🌷🌼
🌸🌺🌻🌹🌷🌼
   🌈 DIVERSITY 🌈
   🏳️‍🌈🏳️‍🌈🏳️‍🌈`,
	}
	
	artIndex := i.rand.Intn(len(asciiArts))
	asciiArt := asciiArts[artIndex]
	
	result := fmt.Sprintf("🎨 Сгенерирован случайный ASCII-арт\n")
	result += fmt.Sprintf("📝 Промпт: %s\n", prompt)
	result += fmt.Sprintf("📐 Размер: %dx%d\n\n", width, height)
	result += asciiArt
	result += "\n\n💡 Для генерации реальных изображений установите API ключ: setSDKey('your_key')"
	result += "\n🔑 Получить ключ: https://platform.stability.ai/account/keys"
	
	if !strings.HasSuffix(filename, ".txt") {
		filename = filename + ".txt"
	}
	
	if err := os.WriteFile(filename, []byte(result), 0644); err != nil {
		return NewTypedString(result), nil
	}
	
	return NewTypedString(result + fmt.Sprintf("\n\n📁 Сохранено в файл: %s", filename)), nil
}

func (i *Interpreter) getLGBTImageHistory(args []TypedValue) (TypedValue, error) {
	files, err := filepath.Glob("lgbt_image_*.png")
	if err != nil {
		return TypedValue{}, err
	}
	
	asciiFiles, _ := filepath.Glob("lgbt_ascii_*.txt")
	allFiles := append(files, asciiFiles...)
	
	if len(allFiles) == 0 {
		return NewTypedString("📂 Нет сгенерированных изображений"), nil
	}
	
	result := fmt.Sprintf("📂 Сгенерированные изображения (%d):\n", len(allFiles))
	for _, f := range allFiles {
		info, _ := os.Stat(f)
		size := info.Size() / 1024
		result += fmt.Sprintf("  • %s (%d KB)\n", f, size)
	}
	
	return NewTypedString(result), nil
}

func (i *Interpreter) deleteLGBTImage(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("deleteLGBTImage: expected filename")
	}
	filename, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("deleteLGBTImage: first argument must be string")
	}
	
	if err := os.Remove(filename); err != nil {
		return TypedValue{}, fmt.Errorf("error deleting file: %v", err)
	}
	
	return NewTypedString(fmt.Sprintf("✅ Файл '%s' удален", filename)), nil
}

// ============================================
// ФУНКЦИЯ createFile - создание файлов
// ============================================

func (i *Interpreter) createFile(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("createFile: expected filename")
	}
	
	filename, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("createFile: first argument must be string")
	}
	
	if err := i.sandbox.CheckFilePath(filename); err != nil {
		return TypedValue{}, err
	}
	
	content := ""
	if len(args) > 1 {
		if c, ok := args[1].Value.(string); ok {
			content = c
		} else {
			return TypedValue{}, fmt.Errorf("createFile: second argument must be string")
		}
	}
	
	mode := "write"
	if len(args) > 2 {
		if m, ok := args[2].Value.(string); ok {
			if m == "append" || m == "write" {
				mode = m
			} else {
				return TypedValue{}, fmt.Errorf("createFile: mode must be 'write' or 'append'")
			}
		}
	}
	
	if int64(len(content)) > i.sandbox.maxFileSize {
		return TypedValue{}, fmt.Errorf("file content too large: %d bytes (max: %d)", len(content), i.sandbox.maxFileSize)
	}
	
	var err error
	if mode == "append" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return TypedValue{}, fmt.Errorf("error opening file for append: %v", err)
		}
		defer f.Close()
		
		_, err = f.WriteString(content)
		if err != nil {
			return TypedValue{}, fmt.Errorf("error appending to file: %v", err)
		}
	} else {
		err = os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return TypedValue{}, fmt.Errorf("error writing file: %v", err)
		}
	}
	
	info, err := os.Stat(filename)
	if err != nil {
		return NewTypedString(fmt.Sprintf("✅ Файл '%s' создан (размер: %d байт)", filename, len(content))), nil
	}
	
	result := fmt.Sprintf("✅ Файл '%s' создан\n", filename)
	result += fmt.Sprintf("📁 Путь: %s\n", filename)
	result += fmt.Sprintf("📄 Размер: %d байт\n", len(content))
	result += fmt.Sprintf("🔄 Режим: %s\n", mode)
	result += fmt.Sprintf("📅 Время создания: %s", info.ModTime().Format("2006-01-02 15:04:05"))
	
	return NewTypedString(result), nil
}

// ============================================
// ФУНКЦИЯ runProgram - запуск внешних программ
// ============================================

func (i *Interpreter) runProgram(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("runProgram: expected command")
	}
	
	command, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("runProgram: first argument must be string")
	}
	
	dangerousCommands := []string{
		"rm", "del", "format", "mkfs", "dd", "shutdown", "reboot",
		"systemctl", "service", "init", "poweroff", "halt",
	}
	
	cmdLower := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(cmdLower, dangerous) {
			return TypedValue{}, fmt.Errorf("runProgram: command '%s' is blocked for security reasons", command)
		}
	}
	
	var argsList []string
	if len(args) > 1 {
		if arr, ok := args[1].Value.([]TypedValue); ok {
			for _, arg := range arr {
				if arg.Type == TypeString {
					argsList = append(argsList, arg.Value.(string))
				} else {
					argsList = append(argsList, arg.String())
				}
			}
		} else if argStr, ok := args[1].Value.(string); ok {
			argsList = strings.Fields(argStr)
		} else {
			return TypedValue{}, fmt.Errorf("runProgram: second argument must be array or string")
		}
	}
	
	workDir := ""
	if len(args) > 2 {
		if dir, ok := args[2].Value.(string); ok {
			if err := i.sandbox.CheckFilePath(dir); err != nil {
				return TypedValue{}, err
			}
			workDir = dir
		}
	}
	
	timeout := 30
	if len(args) > 3 {
		if t, ok := args[3].Value.(int); ok {
			if t > 0 && t <= 300 {
				timeout = t
			}
		}
	}
	
	var cmd *exec.Cmd
	if len(argsList) > 0 {
		cmd = exec.Command(command, argsList...)
	} else {
		cmd = exec.Command(command)
	}
	
	if workDir != "" {
		cmd.Dir = workDir
	}
	
	cmd.Stdout = output
	cmd.Stderr = output
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return TypedValue{}, fmt.Errorf("runProgram: command failed: %v", err)
		}
		result := fmt.Sprintf("✅ Команда '%s' выполнена успешно\n", command)
		if len(argsList) > 0 {
			result += fmt.Sprintf("📝 Аргументы: %v\n", argsList)
		}
		if workDir != "" {
			result += fmt.Sprintf("📁 Рабочая директория: %s\n", workDir)
		}
		result += fmt.Sprintf("⏱️ Таймаут: %d сек", timeout)
		return NewTypedString(result), nil
		
	case <-ctx.Done():
		return TypedValue{}, fmt.Errorf("runProgram: command timed out after %d seconds", timeout)
	}
}

func (i *Interpreter) fileInfo(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("fileInfo: expected filename")
	}
	
	filename, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("fileInfo: first argument must be string")
	}
	
	if err := i.sandbox.CheckFilePath(filename); err != nil {
		return TypedValue{}, err
	}
	
	info, err := os.Stat(filename)
	if err != nil {
		return TypedValue{}, fmt.Errorf("fileInfo: %v", err)
	}
	
	result := fmt.Sprintf("📁 Информация о файле '%s':\n", filename)
	result += fmt.Sprintf("📄 Имя: %s\n", info.Name())
	result += fmt.Sprintf("📏 Размер: %d байт (%d KB)\n", info.Size(), info.Size()/1024)
	result += fmt.Sprintf("📅 Изменен: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("🔐 Права: %v\n", info.Mode())
	result += fmt.Sprintf("📂 Директория: %v", info.IsDir())
	
	return NewTypedString(result), nil
}

func (i *Interpreter) copyFile(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("copyFile: expected source and destination")
	}
	
	src, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("copyFile: first argument must be string")
	}
	
	dst, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("copyFile: second argument must be string")
	}
	
	if err := i.sandbox.CheckFilePath(src); err != nil {
		return TypedValue{}, err
	}
	if err := i.sandbox.CheckFilePath(dst); err != nil {
		return TypedValue{}, err
	}
	
	srcFile, err := os.Open(src)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error opening source: %v", err)
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error creating destination: %v", err)
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error copying: %v", err)
	}
	
	return NewTypedString(fmt.Sprintf("✅ Файл скопирован: %s -> %s", src, dst)), nil
}

func (i *Interpreter) deleteFile(args []TypedValue) (TypedValue, error) {
	if len(args) < 1 {
		return TypedValue{}, fmt.Errorf("deleteFile: expected filename")
	}
	
	filename, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("deleteFile: first argument must be string")
	}
	
	if err := i.sandbox.CheckFilePath(filename); err != nil {
		return TypedValue{}, err
	}
	
	if _, err := os.Stat(filename); err != nil {
		return TypedValue{}, fmt.Errorf("deleteFile: file not found: %v", err)
	}
	
	err := os.Remove(filename)
	if err != nil {
		return TypedValue{}, fmt.Errorf("error deleting file: %v", err)
	}
	
	return NewTypedString(fmt.Sprintf("✅ Файл '%s' удален", filename)), nil
}

// ============================================
// ВСТРОЕННЫЕ ФУНКЦИИ
// ============================================

func (i *Interpreter) getBuiltinFunction(name string) (func([]TypedValue) (TypedValue, error), bool) {
	builtins := map[string]func([]TypedValue) (TypedValue, error){
		"readFile": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("readFile: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("readFile: argument must be string")
			}
			filename := args[0].Value.(string)
			
			if err := i.sandbox.CheckFilePath(filename); err != nil {
				return TypedValue{}, err
			}
			
			info, err := os.Stat(filename)
			if err != nil {
				return TypedValue{}, err
			}
			if info.Size() > i.sandbox.maxFileSize {
				return TypedValue{}, fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), i.sandbox.maxFileSize)
			}
			
			data, err := os.ReadFile(filename)
			if err != nil {
				return TypedValue{}, err
			}
			return NewTypedString(string(data)), nil
		},
		"writeFile": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("writeFile: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("writeFile: first argument must be string")
			}
			if args[1].Type != TypeString {
				return TypedValue{}, fmt.Errorf("writeFile: second argument must be string")
			}
			filename := args[0].Value.(string)
			
			if err := i.sandbox.CheckFilePath(filename); err != nil {
				return TypedValue{}, err
			}
			
			content := args[1].Value.(string)
			if int64(len(content)) > i.sandbox.maxFileSize {
				return TypedValue{}, fmt.Errorf("content too large: %d bytes (max: %d)", len(content), i.sandbox.maxFileSize)
			}
			
			err := os.WriteFile(filename, []byte(content), 0644)
			return TypedValue{Type: TypeNull, Value: nil}, err
		},
		"fileExists": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("fileExists: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("fileExists: argument must be string")
			}
			filename := args[0].Value.(string)
			_, err := os.Stat(filename)
			return NewTypedBool(err == nil), nil
		},
		"getDirFiles": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("getDirFiles: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("getDirFiles: argument must be string")
			}
			dir := args[0].Value.(string)
			
			if err := i.sandbox.CheckFilePath(dir); err != nil {
				return TypedValue{}, err
			}
			
			entries, err := os.ReadDir(dir)
			if err != nil {
				return TypedValue{}, err
			}
			files := make([]TypedValue, len(entries))
			for idx, entry := range entries {
				files[idx] = NewTypedString(entry.Name())
			}
			return NewTypedArray(files), nil
		},
		"split": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("split: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("split: first argument must be string")
			}
			if args[1].Type != TypeString {
				return TypedValue{}, fmt.Errorf("split: second argument must be string")
			}
			text := args[0].Value.(string)
			delim := args[1].Value.(string)
			parts := strings.Split(text, delim)
			result := make([]TypedValue, len(parts))
			for i, p := range parts {
				result[i] = NewTypedString(p)
			}
			return NewTypedArray(result), nil
		},
		"replace": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 3 {
				return TypedValue{}, fmt.Errorf("replace: expected 3 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString || args[1].Type != TypeString || args[2].Type != TypeString {
				return TypedValue{}, fmt.Errorf("replace: all arguments must be strings")
			}
			text := args[0].Value.(string)
			old := args[1].Value.(string)
			new := args[2].Value.(string)
			return NewTypedString(strings.ReplaceAll(text, old, new)), nil
		},
		"trim": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("trim: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("trim: argument must be string")
			}
			text := args[0].Value.(string)
			return NewTypedString(strings.TrimSpace(text)), nil
		},
		"length": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("length: expected 1 argument, got %d", len(args))
			}
			switch args[0].Type {
			case TypeString:
				return NewTypedInt(len(args[0].Value.(string))), nil
			case TypeArray:
				return NewTypedInt(len(args[0].Value.([]TypedValue))), nil
			default:
				return TypedValue{}, fmt.Errorf("length: argument must be string or array")
			}
		},
		"toUpper": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("toUpper: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("toUpper: argument must be string")
			}
			text := args[0].Value.(string)
			return NewTypedString(strings.ToUpper(text)), nil
		},
		"toLower": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("toLower: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("toLower: argument must be string")
			}
			text := args[0].Value.(string)
			return NewTypedString(strings.ToLower(text)), nil
		},
		"append": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("append: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeArray {
				return TypedValue{}, fmt.Errorf("append: first argument must be array")
			}
			arr := args[0].Value.([]TypedValue)
			return NewTypedArray(append(arr, args[1])), nil
		},
		"remove": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("remove: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeArray {
				return TypedValue{}, fmt.Errorf("remove: first argument must be array")
			}
			if args[1].Type != TypeInt {
				return TypedValue{}, fmt.Errorf("remove: second argument must be integer")
			}
			arr := args[0].Value.([]TypedValue)
			idx := args[1].Value.(int)
			if idx < 0 || idx >= len(arr) {
				return TypedValue{}, fmt.Errorf("remove: index out of range")
			}
			return NewTypedArray(append(arr[:idx], arr[idx+1:]...)), nil
		},
		"random": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("random: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeInt || args[1].Type != TypeInt {
				return TypedValue{}, fmt.Errorf("random: arguments must be integers")
			}
			min := args[0].Value.(int)
			max := args[1].Value.(int)
			if min > max {
				return TypedValue{}, fmt.Errorf("random: min must be <= max")
			}
			result := min + i.rand.Intn(max-min+1)
			return NewTypedInt(result), nil
		},
		"max": func(args []TypedValue) (TypedValue, error) {
			if len(args) < 2 {
				return TypedValue{}, fmt.Errorf("max: expected at least 2 arguments")
			}
			maxVal := args[0]
			for _, arg := range args[1:] {
				cmp, err := compareTypedValues(arg, maxVal)
				if err != nil {
					return TypedValue{}, err
				}
				if cmp > 0 {
					maxVal = arg
				}
			}
			return maxVal, nil
		},
		"min": func(args []TypedValue) (TypedValue, error) {
			if len(args) < 2 {
				return TypedValue{}, fmt.Errorf("min: expected at least 2 arguments")
			}
			minVal := args[0]
			for _, arg := range args[1:] {
				cmp, err := compareTypedValues(arg, minVal)
				if err != nil {
					return TypedValue{}, err
				}
				if cmp < 0 {
					minVal = arg
				}
			}
			return minVal, nil
		},
		"sqrt": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("sqrt: expected 1 argument, got %d", len(args))
			}
			var val float64
			switch args[0].Type {
			case TypeInt:
				val = float64(args[0].Value.(int))
			case TypeFloat:
				val = args[0].Value.(float64)
			default:
				return TypedValue{}, fmt.Errorf("sqrt: argument must be number")
			}
			if val < 0 {
				return TypedValue{}, fmt.Errorf("sqrt: cannot take square root of negative number")
			}
			result := val / 2
			for i := 0; i < 100; i++ {
				result = (result + val/result) / 2
			}
			return NewTypedFloat(result), nil
		},
		"pow": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("pow: expected 2 arguments, got %d", len(args))
			}
			var base float64
			switch args[0].Type {
			case TypeInt:
				base = float64(args[0].Value.(int))
			case TypeFloat:
				base = args[0].Value.(float64)
			default:
				return TypedValue{}, fmt.Errorf("pow: first argument must be number")
			}
			var exp float64
			switch args[1].Type {
			case TypeInt:
				exp = float64(args[1].Value.(int))
			case TypeFloat:
				exp = args[1].Value.(float64)
			default:
				return TypedValue{}, fmt.Errorf("pow: second argument must be number")
			}
			if exp == float64(int(exp)) {
				result := 1.0
				for i := 0; i < int(exp); i++ {
					result *= base
				}
				return NewTypedFloat(result), nil
			}
			return NewTypedFloat(float64(1)), nil
		},
		"getTime": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getTime: expected 0 arguments")
			}
			return NewTypedString(time.Now().Format("2006-01-02 15:04:05")), nil
		},
		"getYear": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getYear: expected 0 arguments")
			}
			return NewTypedInt(time.Now().Year()), nil
		},
		"getMonth": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getMonth: expected 0 arguments")
			}
			return NewTypedInt(int(time.Now().Month())), nil
		},
		"getOS": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getOS: expected 0 arguments")
			}
			return NewTypedString(os.Getenv("OS")), nil
		},
		"getHostname": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getHostname: expected 0 arguments")
			}
			hostname, err := os.Hostname()
			if err != nil {
				return TypedValue{}, err
			}
			return NewTypedString(hostname), nil
		},
		"getArgs": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("getArgs: expected 0 arguments")
			}
			argsSlice := os.Args[1:]
			result := make([]TypedValue, len(argsSlice))
			for i, a := range argsSlice {
				result[i] = NewTypedString(a)
			}
			return NewTypedArray(result), nil
		},
		"hasFlag": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("hasFlag: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("hasFlag: argument must be string")
			}
			flag := args[0].Value.(string)
			for _, arg := range os.Args[1:] {
				if arg == flag {
					return NewTypedBool(true), nil
				}
			}
			return NewTypedBool(false), nil
		},
		"httpGet": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("httpGet: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("httpGet: argument must be string")
			}
			url := args[0].Value.(string)
			
			if err := i.sandbox.CheckURL(url); err != nil {
				return TypedValue{}, err
			}
			
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				return TypedValue{}, err
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				return TypedValue{}, fmt.Errorf("HTTP error: %s", resp.Status)
			}
			
			body, err := io.ReadAll(io.LimitReader(resp.Body, i.sandbox.maxFileSize))
			if err != nil {
				return TypedValue{}, err
			}
			return NewTypedString(string(body)), nil
		},
		"httpPost": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("httpPost: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString || args[1].Type != TypeString {
				return TypedValue{}, fmt.Errorf("httpPost: arguments must be strings")
			}
			url := args[0].Value.(string)
			data := args[1].Value.(string)
			
			if err := i.sandbox.CheckURL(url); err != nil {
				return TypedValue{}, err
			}
			
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Post(url, "application/json", strings.NewReader(data))
			if err != nil {
				return TypedValue{}, err
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				return TypedValue{}, fmt.Errorf("HTTP error: %s", resp.Status)
			}
			
			body, err := io.ReadAll(io.LimitReader(resp.Body, i.sandbox.maxFileSize))
			if err != nil {
				return TypedValue{}, err
			}
			return NewTypedString(string(body)), nil
		},
		"jsonParse": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("jsonParse: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("jsonParse: argument must be string")
			}
			jsonStr := args[0].Value.(string)
			
			if int64(len(jsonStr)) > i.sandbox.maxFileSize {
				return TypedValue{}, fmt.Errorf("JSON too large: %d bytes", len(jsonStr))
			}
			
			var result interface{}
			err := json.Unmarshal([]byte(jsonStr), &result)
			if err != nil {
				return TypedValue{}, fmt.Errorf("invalid JSON: %v", err)
			}
			
			return i.jsonToTypedValue(result), nil
		},
		"md5": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("md5: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("md5: argument must be string")
			}
			text := args[0].Value.(string)
			hash := md5.Sum([]byte(text))
			return NewTypedString(hex.EncodeToString(hash[:])), nil
		},
		"sha256": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("sha256: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeString {
				return TypedValue{}, fmt.Errorf("sha256: argument must be string")
			}
			text := args[0].Value.(string)
			hash := sha256.Sum256([]byte(text))
			return NewTypedString(hex.EncodeToString(hash[:])), nil
		},
		"regexFind": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 2 {
				return TypedValue{}, fmt.Errorf("regexFind: expected 2 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString || args[1].Type != TypeString {
				return TypedValue{}, fmt.Errorf("regexFind: arguments must be strings")
			}
			pattern := args[0].Value.(string)
			text := args[1].Value.(string)
			re, err := regexp.Compile(pattern)
			if err != nil {
				return TypedValue{}, err
			}
			matches := re.FindAllString(text, -1)
			result := make([]TypedValue, len(matches))
			for i, m := range matches {
				result[i] = NewTypedString(m)
			}
			return NewTypedArray(result), nil
		},
		"regexReplace": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 3 {
				return TypedValue{}, fmt.Errorf("regexReplace: expected 3 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString || args[1].Type != TypeString || args[2].Type != TypeString {
				return TypedValue{}, fmt.Errorf("regexReplace: all arguments must be strings")
			}
			pattern := args[0].Value.(string)
			text := args[1].Value.(string)
			replacement := args[2].Value.(string)
			re, err := regexp.Compile(pattern)
			if err != nil {
				return TypedValue{}, err
			}
			return NewTypedString(re.ReplaceAllString(text, replacement)), nil
		},
		"sleep": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 1 {
				return TypedValue{}, fmt.Errorf("sleep: expected 1 argument, got %d", len(args))
			}
			if args[0].Type != TypeInt {
				return TypedValue{}, fmt.Errorf("sleep: argument must be integer")
			}
			ms := args[0].Value.(int)
			if ms > 60000 {
				return TypedValue{}, fmt.Errorf("sleep: maximum 60000ms")
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return TypedValue{Type: TypeNull, Value: nil}, nil
		},
		"sendEmail": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 3 {
				return TypedValue{}, fmt.Errorf("sendEmail: expected 3 arguments, got %d", len(args))
			}
			if args[0].Type != TypeString || args[1].Type != TypeString || args[2].Type != TypeString {
				return TypedValue{}, fmt.Errorf("sendEmail: all arguments must be strings")
			}
			to := args[0].Value.(string)
			subject := args[1].Value.(string)
			body := args[2].Value.(string)
			fmt.Fprintf(output, "Email sent to %s: %s - %s\n", to, subject, body)
			return TypedValue{Type: TypeNull, Value: nil}, nil
		},
		"flag": func(args []TypedValue) (TypedValue, error) {
			if len(args) != 0 {
				return TypedValue{}, fmt.Errorf("flag: expected 0 arguments, got %d", len(args))
			}
			
			rainbowColors := []string{
				"#FF0000", "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#4B0082", "#8B00FF",
			}
			
			result := make([]TypedValue, len(rainbowColors))
			for i, color := range rainbowColors {
				result[i] = NewTypedString(color)
			}
			return NewTypedArray(result), nil
		},
		"createServer":          i.createServer,
		"startServer":           i.startServer,
		"stopServer":            i.stopServer,
		"addRoute":              i.addRoute,
		"getServerStatus":       i.getServerStatus,
		"listServers":           i.listServers,
		"antiHomoPhobe":         i.antiHomoPhobe,
		"setSDKey":              i.setSDKey,
		"setSDModel":            i.setSDModel,
		"generateLGBTImage":     i.generateLGBTImage,
		"getLGBTImageHistory":   i.getLGBTImageHistory,
		"deleteLGBTImage":       i.deleteLGBTImage,
		"createFile":            i.createFile,
		"runProgram":            i.runProgram,
		"fileInfo":              i.fileInfo,
		"copyFile":              i.copyFile,
		"deleteFile":            i.deleteFile,
	}

	fn, ok := builtins[name]
	return fn, ok
}

func (i *Interpreter) jsonToTypedValue(v interface{}) TypedValue {
	if v == nil {
		return TypedValue{Type: TypeNull, Value: nil}
	}

	switch val := v.(type) {
	case string:
		return NewTypedString(val)
	case float64:
		if val == float64(int(val)) {
			return NewTypedInt(int(val))
		}
		return NewTypedFloat(val)
	case bool:
		return NewTypedBool(val)
	case []interface{}:
		arr := make([]TypedValue, len(val))
		for idx, elem := range val {
			arr[idx] = i.jsonToTypedValue(elem)
		}
		return NewTypedArray(arr)
	case map[string]interface{}:
		arr := make([]TypedValue, 0, len(val)*2)
		for k, v2 := range val {
			arr = append(arr, NewTypedString(k))
			arr = append(arr, i.jsonToTypedValue(v2))
		}
		return NewTypedArray(arr)
	default:
		return NewTypedString(fmt.Sprintf("%v", v))
	}
}

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

func (i *Interpreter) handleInclude(n *IncludeStatement) (TypedValue, error) {
	filename := n.Filename

	if err := i.sandbox.CheckFilePath(filename); err != nil {
		return TypedValue{}, err
	}

	searchPaths := []string{
		".",
	}

	if currentFilePath != "" {
		dir := filepath.Dir(currentFilePath)
		searchPaths = append(searchPaths, dir)
	}

	searchPaths = append(searchPaths, "libs")
	searchPaths = append(searchPaths, "libraries")

	var fileContent []byte
	var err error
	var found bool

	for _, path := range searchPaths {
		fullPath := filepath.Join(path, filename)
		if !strings.HasSuffix(fullPath, ".rb") && !strings.HasSuffix(fullPath, ".rainbow") {
			testPath := fullPath + ".rb"
			fileContent, err = os.ReadFile(testPath)
			if err == nil {
				found = true
				break
			}
			testPath = fullPath + ".rainbow"
			fileContent, err = os.ReadFile(testPath)
			if err == nil {
				found = true
				break
			}
		} else {
			fileContent, err = os.ReadFile(fullPath)
			if err == nil {
				found = true
				break
			}
		}
	}

	if !found {
		return TypedValue{}, fmt.Errorf("library not found: %s (searched in: %v)", filename, searchPaths)
	}

	if int64(len(fileContent)) > i.sandbox.maxFileSize {
		return TypedValue{}, fmt.Errorf("library too large: %d bytes", len(fileContent))
	}

	oldPath := currentFilePath
	if currentFilePath == "" {
		currentFilePath = filename
	}

	lexer := NewLexer(string(fileContent))
	tokens, err := lexer.Tokenize()
	if err != nil {
		currentFilePath = oldPath
		return TypedValue{}, fmt.Errorf("error tokenizing library %s: %v", filename, err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		currentFilePath = oldPath
		return TypedValue{}, fmt.Errorf("error parsing library %s: %v", filename, err)
	}

	_, err = i.EvaluateProgram(ast)
	if err != nil {
		currentFilePath = oldPath
		return TypedValue{}, fmt.Errorf("error executing library %s: %v", filename, err)
	}

	currentFilePath = oldPath
	return TypedValue{Type: TypeNull, Value: nil}, nil
}

func (i *Interpreter) EvaluateProgram(program *Program) (TypedValue, error) {
	for _, stmt := range program.Statements {
		_, err := i.Evaluate(stmt)
		if err != nil {
			return TypedValue{}, err
		}
	}
	return TypedValue{Type: TypeNull, Value: nil}, nil
}

func (i *Interpreter) Evaluate(node Node) (TypedValue, error) {
	if program, ok := node.(*Program); ok {
		return i.EvaluateProgram(program)
	}

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
			val, err := i.Evaluate(elem)
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
		idxVal, err := i.Evaluate(n.Index)
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
		val, err := i.Evaluate(n.Expr)
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
		default:
			return TypedValue{}, fmt.Errorf("unknown unary operator: %s", n.Op)
		}
	case *BinaryOpNode:
		left, err := i.Evaluate(n.Left)
		if err != nil {
			return TypedValue{}, err
		}
		right, err := i.Evaluate(n.Right)
		if err != nil {
			return TypedValue{}, err
		}
		return i.evaluateBinaryOp(left, n.Op, right)
	case *TypedDeclaration:
		value, err := i.Evaluate(n.Value)
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
		value, err := i.Evaluate(n.Value)
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
		value, err := i.Evaluate(n.Value)
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
			i.setVar(n.Name, result)
			return current, nil
		} else {
			i.setVar(n.Name, result)
			return result, nil
		}
	case *ArrayAssignmentStatement:
		if i.isConstant(n.Name) {
			return TypedValue{}, fmt.Errorf("cannot assign to constant array '%s'", n.Name)
		}
		value, err := i.Evaluate(n.Value)
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
		idxVal, err := i.Evaluate(n.Index)
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
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		fmt.Fprintln(output, value.String())
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *IfStatement:
		cond, err := i.Evaluate(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}
		if i.isTruthy(cond) {
			for _, stmt := range n.ThenBlock {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return TypedValue{}, err
				}
			}
			return TypedValue{Type: TypeNull, Value: nil}, nil
		}

		for _, elseIf := range n.ElseIfBlocks {
			cond, err := i.Evaluate(elseIf.Condition)
			if err != nil {
				return TypedValue{}, err
			}
			if i.isTruthy(cond) {
				for _, stmt := range elseIf.Block {
					_, err := i.Evaluate(stmt)
					if err != nil {
						return TypedValue{}, err
					}
				}
				return TypedValue{Type: TypeNull, Value: nil}, nil
			}
		}

		for _, stmt := range n.ElseBlock {
			_, err := i.Evaluate(stmt)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *WhileStatement:
		condVal, err := i.Evaluate(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}
		for i.isTruthy(condVal) {
			for _, stmt := range n.Body {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return TypedValue{}, err
				}
				if i.returnFlag {
					return i.returnValue, nil
				}
			}
			condVal, err = i.Evaluate(n.Condition)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *SexStatement:
		if n.Init != nil {
			_, err := i.Evaluate(n.Init)
			if err != nil {
				return TypedValue{}, err
			}
		}
		
		condVal, err := i.Evaluate(n.Condition)
		if err != nil {
			return TypedValue{}, err
		}
		
		for i.isTruthy(condVal) {
			for _, stmt := range n.Body {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return TypedValue{}, err
				}
				if i.returnFlag {
					return i.returnValue, nil
				}
			}
			
			if n.Update != nil {
				_, err = i.Evaluate(n.Update)
				if err != nil {
					return TypedValue{}, err
				}
			}
			
			condVal, err = i.Evaluate(n.Condition)
			if err != nil {
				return TypedValue{}, err
			}
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *ConstantDeclaration:
		value, err := i.Evaluate(n.Value)
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
		countryVal, err := i.Evaluate(n.Country)
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
		if fn, ok := i.getBuiltinFunction(n.Name); ok {
			args := make([]TypedValue, len(n.Args))
			for idx, arg := range n.Args {
				val, err := i.Evaluate(arg)
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
			val, err := i.Evaluate(arg)
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
			lastResult, err = i.Evaluate(stmt)
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
	case *ReturnStatement:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return TypedValue{}, err
		}
		i.returnValue = value
		i.returnFlag = true
		return value, nil
	case *TryCatchStatement:
		var err error
		var caughtErr error
		for _, stmt := range n.TryBlock {
			_, err = i.Evaluate(stmt)
			if err != nil {
				caughtErr = err
				break
			}
		}

		if err != nil {
			line := 0
			col := 0
			if n, ok := node.(Node); ok {
				line = n.GetLine()
			}
			i.setVar("error", NewTypedString(caughtErr.Error()))
			i.setType("error", "lesbian")
			for _, stmt := range n.CatchBlock {
				_, catchErr := i.Evaluate(stmt)
				if catchErr != nil {
					return TypedValue{}, catchErr
				}
			}
			i.handleError(caughtErr, line, col)
			return TypedValue{Type: TypeNull, Value: nil}, nil
		}
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *ExpressionStatement:
		val, err := i.Evaluate(n.Expr)
		if err != nil {
			return TypedValue{}, err
		}
		return val, nil
	default:
		return TypedValue{Type: TypeNull, Value: nil}, nil
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

func (i *Interpreter) showHelp(country string) {
	orgs := map[string][]string{
		"russia": {
			"Российская ЛГБТ-сеть",
			"Центр 'Сфера'",
			"Проект 'ЛГБТ-Киров'",
			"Выход (Санкт-Петербург)",
		},
		"usa": {
			"Human Rights Campaign",
			"GLAAD",
			"The Trevor Project",
			"National LGBTQ Task Force",
		},
		"all": {
			"ILGA (International Lesbian, Gay, Bisexual, Trans and Intersex Association)",
			"ILGA-Europe",
			"OutRight Action International",
		},
	}
	countryLower := strings.ToLower(country)
	if list, ok := orgs[countryLower]; ok {
		fmt.Fprintf(output, "Организации, помогающие ЛГБТ в стране '%s':\n", country)
		for _, org := range list {
			fmt.Fprintf(output, "  • %s\n", org)
		}
	} else if countryLower == "" || countryLower == "all" {
		fmt.Fprintln(output, "Международные организации, помогающие ЛГБТ по всему миру:")
		for _, org := range orgs["all"] {
			fmt.Fprintf(output, "  • %s\n", org)
		}
	} else {
		fmt.Fprintf(output, "Информация для страны '%s' не найдена.\n", country)
	}
}

func (i *Interpreter) runOrientationTest() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprintln(output, "\n=== Психологический тест на определение сексуальной ориентации ===")
	fmt.Fprintln(output, "Ответьте на вопросы, выбрав вариант (1-4). Будьте честны.")
	fmt.Fprintln(output, "------------------------------------------------------------")

	questions := []struct {
		text    string
		options []string
		scores  []int
	}{
		{
			"Кто вам чаще нравится в романтическом плане?",
			[]string{"Люди противоположного пола", "Люди любого пола", "Люди того же пола", "Затрудняюсь ответить"},
			[]int{1, 2, 3, 0},
		},
		{
			"Какие отношения вы считаете наиболее привлекательными для себя?",
			[]string{"Гетеросексуальные", "Бисексуальные", "Гомосексуальные", "Не уверен"},
			[]int{1, 2, 3, 0},
		},
		{
			"Как часто вы испытывали романтические чувства к людям своего пола?",
			[]string{"Никогда", "Иногда", "Часто", "Постоянно"},
			[]int{1, 2, 3, 4},
		},
		{
			"Как часто вы испытывали романтические чувства к людям противоположного пола?",
			[]string{"Постоянно", "Часто", "Иногда", "Никогда"},
			[]int{1, 2, 3, 4},
		},
		{
			"Считаете ли вы, что ориентация может меняться со временем?",
			[]string{"Нет, это врождённое", "Возможно, но редко", "Да, это гибко", "Затрудняюсь ответить"},
			[]int{1, 2, 3, 0},
		},
	}

	totalScore := 0
	for idx, q := range questions {
		fmt.Fprintf(output, "\n%d. %s\n", idx+1, q.text)
		for j, opt := range q.options {
			fmt.Fprintf(output, "   %d) %s\n", j+1, opt)
		}
		fmt.Fprint(output, "Ваш выбор (1-4): ")
		scanner.Scan()
		input := scanner.Text()
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > 4 {
			fmt.Fprintln(output, "Некорректный ввод, принимаем 1.")
			choice = 1
		}
		totalScore += q.scores[choice-1]
	}

	var result string
	if totalScore >= 12 {
		result = "Гомосексуальная ориентация (выраженная)"
	} else if totalScore >= 8 {
		result = "Бисексуальная ориентация"
	} else if totalScore >= 4 {
		result = "Гетеросексуальная ориентация"
	} else {
		result = "Скорее всего, вы ещё определяетесь или ваша ориентация не укладывается в простые категории."
	}
	fmt.Fprintf(output, "\nРезультат теста (баллы: %d): %s\n", totalScore, result)
	fmt.Fprintln(output, "Помните: этот тест не является диагностическим.")
}

func (i *Interpreter) runOrientationDemo() {
	fmt.Fprintln(output, "\n=== Психологический тест на определение сексуальной ориентации (ДЕМО-ВЕРСИЯ) ===")
	fmt.Fprintln(output, "Для прохождения полного теста запустите программу интерактивно.")
	fmt.Fprintln(output, "Результат демо-теста (усреднённый):")
	fmt.Fprintln(output, "Скорее всего, ваша ориентация уникальна и не укладывается в простые категории.")
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

	_, err = interpreter.EvaluateProgram(ast)
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

func runExample() {
	program := `
@ Пример с инкрементом и декрементом

RAINBOW main() {
    COMINGOUT "🌈 Инкремент и декремент:";
    
    GAY counter = 0;
    COMINGOUT "counter = " + counter;
    
    counter++;
    COMINGOUT "counter++ = " + counter;
    
    ++counter;
    COMINGOUT "++counter = " + counter;
    
    counter--;
    COMINGOUT "counter-- = " + counter;
    
    --counter;
    COMINGOUT "--counter = " + counter;
    
    COMINGOUT "";
    COMINGOUT "📊 Составные операции:";
    
    GAY value = 10;
    COMINGOUT "value = " + value;
    
    value += 5;
    COMINGOUT "value += 5 = " + value;
    
    value -= 3;
    COMINGOUT "value -= 3 = " + value;
    
    value *= 2;
    COMINGOUT "value *= 2 = " + value;
    
    value /= 4;
    COMINGOUT "value /= 4 = " + value;
    
    COMINGOUT "";
    COMINGOUT "📊 Цикл с инкрементом:";
    
    SEX (GAY i = 0; i < 5; i++) {
        COMINGOUT "  i = " + i;
    }
    
    COMINGOUT "";
    COMINGOUT "✅ Все операции работают!";
}

main();
`

	fmt.Fprintln(output, "=== LGBTScript с инкрементом и декрементом ===")
	fmt.Fprintln(output, "\n=== Выполнение ===")

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
	_, err = interpreter.EvaluateProgram(ast)
	if err != nil {
		fmt.Fprintln(output, "Ошибка выполнения:", err)
	}
}

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

	fmt.Fprintf(output, "✅ Скомпилировано: %s -> %s (размер: %d байт, скрипт: %d байт)\n",
		inputScript, outputExe, infoSize(dst), len(scriptData))
	return nil
}

func infoSize(f *os.File) int64 {
	info, err := f.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

func runScript(script string, showTokens, showAST bool) error {
    lexer := NewLexer(script)
    
    interpreter := NewInterpreter()
    
    hasOffensive, terms := interpreter.checkForOffensiveTerms(script)
    if hasOffensive {
        fmt.Printf("⚠️ В коде обнаружены потенциально оскорбительные термины: %v\n", terms)
        fmt.Println("Программа будет запущена с предупреждением. Используйте корректную терминологию.")
        interpreter.showSupportMessage()
        os.Exit(1)
        time.Sleep(1 * time.Second)
    }
    
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

    _, err = interpreter.EvaluateProgram(ast)
    if err != nil {
        return fmt.Errorf("ошибка выполнения: %v", err)
    }
    return nil
}

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

	showTokens := flag.Bool("tokens", false, "вывести токены после лексического анализа")
	showAST := flag.Bool("ast", false, "вывести AST после парсинга")
	command := flag.String("c", "", "выполнить код из командной строки")
	lgbtFile := flag.String("lgbt", "", "исполнить файл с кодом")
	debug := flag.Bool("debug", false, "включить режим отладки")
	example := flag.Bool("example", false, "показать пример с инкрементом и декрементом")
	buildFlag := flag.Bool("b", false, "скомпилировать .rainbow в .exe")

	flag.Parse()

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

	fmt.Fprintln(output, "🌈 LGBTScript - Язык программирования с поддержкой ЛГБТ+ сообщества")
	fmt.Fprintln(output, "📖 Используйте --example для демонстрации инкремента и декремента")
	fmt.Fprintln(output, "📁 Укажите файл .rainbow для выполнения")
	fmt.Fprintln(output, "🔧 Для компиляции в .exe используйте -b input.rainbow output.exe")
	runExample()
}