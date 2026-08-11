package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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
			"lesbian": true, "gay": true, "trans": true, "nonbinary": true, "gender": true,
			"comingout": true, "cis": true, "nocis": true,
			"true": true, "false": true,
			"help": true, "orientation": true,
			"rainbow": true,
			"return": true,
			"try": true, "catch": true,
			"export": true,
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

		// Директива #intersex
		if ch == '#' && l.pos+1 < len(l.input) {
			next := l.input[l.pos+1]
			if next >= 'a' && next <= 'z' {
				start := l.pos
				for l.pos < len(l.input) && ((l.input[l.pos] >= 'a' && l.input[l.pos] <= 'z') || l.input[l.pos] == '#') {
					l.pos++
				}
				directive := l.input[start:l.pos]
				if directive == "#intersex" {
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
			(ch == '&' && next == '&') || (ch == '|' && next == '|') {
			value = string(ch) + string(next)
			l.pos += 2
			l.col += 2
			l.tokens = append(l.tokens, Token{TOKEN_OPERATOR, value, startLine, startCol})
			return nil
		}
	}

	operators := "=+-*/%<>=!(){},;"
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

type ArrayAssignmentStatement struct {
	BaseNode
	Name  string
	Index Node
	Value Node
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
		case "lesbian", "gay", "trans", "nonbinary", "gender":
			return p.parseTypedDeclaration()
		case "comingout":
			return p.parsePrintStatement()
		case "cis":
			return p.parseIfStatement()
		case "pride":
			return p.parseWhileStatement()
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
		}
	}

	if token.Type == TOKEN_IDENTIFIER {
		nextToken := p.peekNext()
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

		if nextToken.Value == "=" {
			return p.parseAssignment()
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
	case "trans":
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

func (p *Parser) parseAssignment() (Node, error) {
	token := p.peek()
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
	return &AssignmentStatement{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name.Value,
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

// ---------- Интерпретатор ----------
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

type callFrame struct {
	vars  map[string]TypedValue
	types map[string]string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:      make(map[string]TypedValue),
		variableTypes:  make(map[string]string),
		functions:      make(map[string]*FunctionDeclaration),
		exportedFuncs:  make(map[string]*FunctionDeclaration),
		callStack:      []callFrame{{vars: make(map[string]TypedValue), types: make(map[string]string)}},
		returnValue:    TypedValue{Type: TypeNull, Value: nil},
		returnFlag:     false,
		maxRecursion:   1000,
		recursionDepth: 0,
		sandbox:        NewSandbox(),
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (i *Interpreter) SetErrorHandler(handler func(error, int, int)) {
	i.errorHandler = handler
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
		vars:  make(map[string]TypedValue),
		types: make(map[string]string),
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
		return fmt.Errorf("type mismatch at line: expected %s (%v), got %v", typ, expectedType, value.Type)
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
// НОВЫЕ СОЦИАЛЬНЫЕ И ПОДДЕРЖИВАЮЩИЕ ФУНКЦИИ
// ============================================

// ---------- Ресурсы и поддержка ----------
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

// ---------- Психологическая поддержка ----------
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

// ---------- Образовательные инструменты ----------
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
		"небинарность": "НЕБИНАРНОСТЬ (Non-binary) — гендерная идентичность, которая не вписывается в бинарную систему мужского и женского пола.",
		"бисексуальность": "БИСЕКСУАЛЬНОСТЬ (Bisexuality) — романтическое и/или сексуальное влечение к людям более чем одного пола.",
		"гомосексуальность": "ГОМОСЕКСУАЛЬНОСТЬ (Homosexuality) — романтическое и/или сексуальное влечение к людям того же пола.",
		"трансгендерность": "ТРАНСГЕНДЕРНОСТЬ (Transgender) — состояние, когда гендерная идентичность человека не совпадает с полом при рождении.",
		"гетеросексуальность": "ГЕТЕРОСЕКСУАЛЬНОСТЬ (Heterosexuality) — романтическое и/или сексуальное влечение к людям противоположного пола.",
		"квир": "КВИР (Queer) — зонтичный термин для ЛГБТ+ сообщества, обозначающий несоответствие нормам.",
		"интерсекс": "ИНТЕРСЕКС (Intersex) — люди, рожденные с репродуктивными или половыми характеристиками, не вписывающимися в типичные определения мужского или женского тела.",
	}

	result := fmt.Sprintf("📖 Определение термина '%s' (язык: %s):\n", term, language)
	if def, ok := terms[strings.ToLower(term)]; ok {
		result += "  " + def + "\n"
	} else {
		result += "  ℹ️ Термин не найден. Рекомендуем обратиться к словарю ЛГБТ+ терминов.\n"
	}

	return NewTypedString(result), nil
}

func (i *Interpreter) lgbtHistoryQuiz(args []TypedValue) (TypedValue, error) {
	difficulty := "medium"
	if len(args) > 0 {
		if d, ok := args[0].Value.(string); ok {
			difficulty = d
		}
	}

	questions := []string{
		"1. В каком году произошли Стоунволлские бунты?\n   A) 1969  B) 1975  C) 1980  D) 1990\n   Ответ: A",
		"2. Кто был первой открытой ЛГБТ-персоной в Конгрессе США?\n   A) Харви Милк  B) Барни Фрэнк  C) Тэмми Болдуин  D) Джим Коллинз\n   Ответ: B",
		"3. В каком году однополые браки были легализованы в США?\n   A) 2010  B) 2015  C) 2020  D) 2005\n   Ответ: B",
		"4. Кто из этих людей был ЛГБТ-активистом?\n   A) Мартин Лютер Кинг  B) Харви Милк  C) Нельсон Мандела  D) Махатма Ганди\n   Ответ: B",
		"5. В каком году ВОЗ исключила гомосексуальность из списка психических расстройств?\n   A) 1973  B) 1990  C) 2000  D) 1985\n   Ответ: B",
	}

	idx := i.rand.Intn(len(questions))
	result := fmt.Sprintf("📚 Квиз по истории ЛГБТ+ (сложность: %s):\n\n", difficulty)
	result += questions[idx] + "\n\n"
	result += "💡 Проверьте свои знания и узнавайте больше!"

	return NewTypedString(result), nil
}

func (i *Interpreter) getDailyFact(args []TypedValue) (TypedValue, error) {
	facts := []string{
		"🏳️‍🌈 Первый парад гордости состоялся в Нью-Йорке в 1970 году.",
		"📖 В Древней Греции гомосексуальные отношения были распространены и считались нормой.",
		"🎨 Известный художник Леонардо да Винчи был гомосексуалом.",
		"📚 'Сад' (The Garden) — один из первых ЛГБТ-фильмов 1968 года.",
		"🏛️ В 1969 году Стоунволлские бунты стали поворотным моментом для ЛГБТ-движения.",
		"🌍 Первый в мире ЛГБТ-прайд состоялся в 1970 году.",
		"📖 В 1973 году Американская психиатрическая ассоциация исключила гомосексуальность из списка психических расстройств.",
	}

	idx := i.rand.Intn(len(facts))
	category := "культура"
	region := "global"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			category = c
		}
	}
	if len(args) > 1 {
		if r, ok := args[1].Value.(string); ok {
			region = r
		}
	}

	return NewTypedString(fmt.Sprintf("📌 Факт дня (категория: %s, регион: %s):\n%s", category, region, facts[idx])), nil
}

// ---------- Медицинская информация ----------
func (i *Interpreter) getHRTInfo(args []TypedValue) (TypedValue, error) {
	country := "USA"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	hrtType := "MTF"
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			hrtType = t
		}
	}

	result := fmt.Sprintf("🏥 Информация о гормональной терапии (страна: %s, тип: %s):\n", country, hrtType)
	result += "\n📋 Основная информация:\n"
	result += "  • Гормональная терапия проводится под наблюдением эндокринолога\n"
	result += "  • Требуется регулярная сдача анализов крови\n"
	result += "  • Эффект проявляется в течение 3-6 месяцев\n\n"

	result += "🔍 Рекомендации:\n"
	result += "  • Обратитесь к ЛГБТ-дружественному эндокринологу\n"
	result += "  • Получите направление от психотерапевта (по требованиям страны)\n"
	result += "  • Обсудите все риски и побочные эффекты с врачом\n"

	return NewTypedString(result), nil
}

func (i *Interpreter) findLGBTDoctor(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("findLGBTDoctor: expected specialty and city")
	}
	specialty, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("findLGBTDoctor: first argument must be string")
	}
	city, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("findLGBTDoctor: second argument must be string")
	}

	doctors := []string{
		"👨‍⚕️ Иванова Е.П. (терапевт) - клиника 'Здоровье', ул. Ленина, 10",
		"👩‍⚕️ Петров С.М. (гинеколог) - центр 'Женское здоровье', пр. Мира, 25",
		"👨‍⚕️ Сидорова А.А. (эндокринолог) - клиника 'Гармония', ул. Садовая, 5",
		"👩‍⚕️ Козлова Н.В. (психотерапевт) - центр 'Поддержка', ул. Свободы, 42",
		"👨‍⚕️ Морозов Д.И. (уролог) - клиника 'Здоровье мужчины', ул. Спортивная, 15",
	}

	result := fmt.Sprintf("🏥 ЛГБТ-дружественные врачи (%s) в городе %s:\n", specialty, city)
	for _, doc := range doctors {
		result += "  • " + doc + "\n"
	}
	result += "\n💡 Уточните информацию о приеме по телефону клиники."

	return NewTypedString(result), nil
}

func (i *Interpreter) getDocumentChangeGuide(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("getDocumentChangeGuide: expected country and document type")
	}
	country, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getDocumentChangeGuide: first argument must be string")
	}
	document, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("getDocumentChangeGuide: second argument must be string")
	}

	guide := map[string]string{
		"паспорт": "Для смены паспорта необходимы: заявление, медицинское заключение, новый паспорт.",
		"свидетельство": "Для смены свидетельства о рождении требуется решение суда.",
		"водительские": "Для смены водительского удостоверения необходимо заявление в ГИБДД.",
	}

	result := fmt.Sprintf("📄 Инструкция по смене документа '%s' в стране '%s':\n", document, country)
	if info, ok := guide[document]; ok {
		result += "  • " + info + "\n"
	} else {
		result += "  • ℹ️ Информация уточняется. Рекомендуем обратиться к юристу.\n"
	}
	result += "\n⚖️ Рекомендуем получить юридическую консультацию."

	return NewTypedString(result), nil
}

// ---------- События и сообщество ----------
func (i *Interpreter) getLGBTQEvents(args []TypedValue) (TypedValue, error) {
	days := 30
	if len(args) > 0 {
		if d, ok := args[0].Value.(int); ok {
			days = d
		}
	}

	eventType := "online"
	if len(args) > 1 {
		if t, ok := args[1].Value.(string); ok {
			eventType = t
		}
	}

	events := []string{
		"🏳️‍🌈 Онлайн-встреча 'Разговор о важном' - 15 марта, 19:00",
		"📚 Книжный клуб 'Радужные страницы' - каждую субботу, 16:00",
		"🎨 Арт-терапия 'Вырази себя' - 20 марта, 18:00",
		"💬 Группа поддержки 'Вместе' - по вторникам, 20:00",
		"🏳️‍⚧️ Транс-встреча 'Голоса' - 25 марта, 19:30",
	}

	result := fmt.Sprintf("📅 Ближайшие ЛГБТ+ мероприятия (следующие %d дней, тип: %s):\n", days, eventType)
	for _, event := range events {
		result += "  • " + event + "\n"
	}
	result += "\n🔗 Подробнее на https://lgbt-events.org"

	return NewTypedString(result), nil
}

func (i *Interpreter) createLGBTQGroup(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("createLGBTQGroup: expected name and meeting type")
	}
	name, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("createLGBTQGroup: first argument must be string")
	}
	meetingType, ok := args[1].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("createLGBTQGroup: second argument must be string")
	}

	result := fmt.Sprintf("🌈 Группа '%s' создана!\n", name)
	result += "📋 Информация о группе:\n"
	result += "  • Тип встреч: " + meetingType + "\n"
	result += "  • Статус: активна\n"
	result += "  • Участников: 0\n\n"
	result += "🔗 Ссылка для присоединения: https://lgbt-groups.org/join/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")

	return NewTypedString(result), nil
}

func (i *Interpreter) findVolunteerOpportunity(args []TypedValue) (TypedValue, error) {
	if len(args) < 2 {
		return TypedValue{}, fmt.Errorf("findVolunteerOpportunity: expected organization and skills")
	}
	organization, ok := args[0].Value.(string)
	if !ok {
		return TypedValue{}, fmt.Errorf("findVolunteerOpportunity: first argument must be string")
	}
	skills, ok := args[1].Value.([]TypedValue)
	if !ok {
		return TypedValue{}, fmt.Errorf("findVolunteerOpportunity: second argument must be array")
	}

	skillsList := make([]string, len(skills))
	for i, s := range skills {
		if str, ok := s.Value.(string); ok {
			skillsList[i] = str
		}
	}

	opportunities := []string{
		"📱 Помощь в ведении социальных сетей",
		"📞 Оператор горячей линии поддержки",
		"📝 Юридическая помощь ЛГБТ+",
		"🎓 Проведение образовательных лекций",
		"🎨 Организация культурных мероприятий",
	}

	result := fmt.Sprintf("🤝 Волонтерские возможности в организации '%s':\n", organization)
	result += "Ваши навыки: " + strings.Join(skillsList, ", ") + "\n\n"
	result += "📋 Доступные позиции:\n"
	for _, opp := range opportunities {
		result += "  • " + opp + "\n"
	}

	return NewTypedString(result), nil
}

// ---------- Культурные и творческие инструменты ----------
func (i *Interpreter) getLGBTQBook(args []TypedValue) (TypedValue, error) {
	genre := "фантастика"
	if len(args) > 0 {
		if g, ok := args[0].Value.(string); ok {
			genre = g
		}
	}

	books := []string{
		"📚 'Имя ветра' - Патрик Ротфусс (фантастика, ЛГБТ+ персонажи)",
		"📚 'Гарри Поттер' - Джоан Роулинг (фэнтези с ЛГБТ+ подтекстом)",
		"📚 'Орландо' - Вирджиния Вулф (классика, гендерная тематика)",
		"📚 'Моррис' - Эдвард Форстер (ЛГБТ+ классика)",
		"📚 'Месяц в глуши' - Эллен Хопкинс (современная проза)",
	}

	idx := i.rand.Intn(len(books))
	return NewTypedString(fmt.Sprintf("📖 Рекомендуемая книга (жанр: %s):\n%s", genre, books[idx])), nil
}

func (i *Interpreter) getLGBTQPlaylist(args []TypedValue) (TypedValue, error) {
	mood := "empowerment"
	if len(args) > 0 {
		if m, ok := args[0].Value.(string); ok {
			mood = m
		}
	}

	playlists := map[string]string{
		"empowerment": "🎵 Плейлист 'Сила и гордость': Lady Gaga, Beyoncé, Madonna, Freddie Mercury",
		"relaxation":  "🎵 Плейлист 'Спокойствие': Enya, Vangelis, Ludovico Einaudi",
		"energy":      "🎵 Плейлист 'Энергия': Daft Punk, The Chemical Brothers, Prodigy",
		"sad":         "🎵 Плейлист 'Меланхолия': Adele, Sam Smith, ЛГБТ-исполнители",
		"love":        "🎵 Плейлист 'Любовь': Elton John, Freddie Mercury, George Michael",
	}

	result := fmt.Sprintf("🎶 Плейлист по настроению '%s':\n", mood)
	if playlist, ok := playlists[mood]; ok {
		result += "  " + playlist + "\n"
	} else {
		result += "  🎵 Рекомендуем: ЛГБТ+ исполнители разных жанров\n"
	}

	return NewTypedString(result), nil
}

func (i *Interpreter) getLGBTQMovies(args []TypedValue) (TypedValue, error) {
	genre := "романтика"
	if len(args) > 0 {
		if g, ok := args[0].Value.(string); ok {
			genre = g
		}
	}

	movies := []string{
		"🎬 'Горбатая гора' (2005) - драма, романтика",
		"🎬 'Кэрол' (2015) - драма, романтика",
		"🎬 'Парень с соседнего кладбища' (2018) - комедия, романтика",
		"🎬 'С тобой или без тебя' (2020) - драма",
		"🎬 'Король Фредерик' (2020) - историческая драма",
	}

	idx := i.rand.Intn(len(movies))
	return NewTypedString(fmt.Sprintf("🎥 Рекомендуемый фильм (жанр: %s):\n%s", genre, movies[idx])), nil
}

// ---------- Встроенные функции ----------
func (i *Interpreter) getBuiltinFunction(name string) (func([]TypedValue) (TypedValue, error), bool) {
	builtins := map[string]func([]TypedValue) (TypedValue, error){
		// Существующие функции
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

		// ---------- НОВЫЕ СОЦИАЛЬНЫЕ ФУНКЦИИ ----------
		"findSafeSpace":       i.findSafeSpace,
		"getCrisisSupport":    i.getCrisisSupport,
		"getLGBTQLaws":        i.getLGBTQLaws,
		"getDailyAffirmation": i.getDailyAffirmation,
		"moodCheck":           i.moodCheck,
		"guidedBreathing":     i.guidedBreathing,
		"defineTerm":          i.defineTerm,
		"lgbtHistoryQuiz":     i.lgbtHistoryQuiz,
		"getDailyFact":        i.getDailyFact,
		"getHRTInfo":          i.getHRTInfo,
		"findLGBTDoctor":      i.findLGBTDoctor,
		"getDocumentChangeGuide": i.getDocumentChangeGuide,
		"getLGBTQEvents":      i.getLGBTQEvents,
		"createLGBTQGroup":    i.createLGBTQGroup,
		"findVolunteerOpportunity": i.findVolunteerOpportunity,
		"getLGBTQBook":        i.getLGBTQBook,
		"getLGBTQPlaylist":    i.getLGBTQPlaylist,
		"getLGBTQMovies":      i.getLGBTQMovies,
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
	case *ArrayAssignmentStatement:
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
			}
			condVal, err = i.Evaluate(n.Condition)
			if err != nil {
				return TypedValue{}, err
			}
		}
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

// ---------- Вспомогательные функции ----------
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
		"uk": {
			"Stonewall",
			"LGBT Foundation",
			"Switchboard (LGBT+ helpline)",
		},
		"germany": {
			"LSVD (Lesben- und Schwulenverband)",
			"CSD Deutschland",
			"Schwulenberatung Berlin",
		},
		"france": {
			"Le Refuge",
			"SOS Homophobie",
			"ACT UP Paris",
		},
		"australia": {
			"ACON",
			"QLife (LGBTQ+ helpline)",
			"Equality Australia",
		},
		"canada": {
			"Egale Canada",
			"Rainbow Health Ontario",
			"PFLAG Canada",
		},
		"all": {
			"ILGA (International Lesbian, Gay, Bisexual, Trans and Intersex Association)",
			"ILGA-Europe",
			"OutRight Action International",
			"Transgender Europe (TGEU)",
			"IGLA (International Gay and Lesbian Association)",
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
		fmt.Fprintln(output, "\nДля получения списка организаций в конкретной стране используйте:")
		fmt.Fprintln(output, `  help "страна" (например, help "russia")`)
	} else {
		fmt.Fprintf(output, "Информация для страны '%s' не найдена. Доступные страны: russia, usa, uk, germany, france, australia, canada.\n", country)
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
	fmt.Fprintln(output, "Помните: этот тест не является диагностическим. Ориентация – это сложный и индивидуальный аспект личности.")
	fmt.Fprintln(output, "Если вы испытываете дискомфорт, обратитесь к психологу или в поддерживающие организации.")
}

func (i *Interpreter) runOrientationDemo() {
	fmt.Fprintln(output, "\n=== Психологический тест на определение сексуальной ориентации (ДЕМО-ВЕРСИЯ) ===")
	fmt.Fprintln(output, "Для прохождения полного теста запустите программу интерактивно.")
	fmt.Fprintln(output, "Пример: rb10.exe")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Вопросы теста:")
	fmt.Fprintln(output, "1. Кто вам чаще нравится в романтическом плане?")
	fmt.Fprintln(output, "   Варианты: противоположный пол, любой пол, тот же пол, не уверен")
	fmt.Fprintln(output, "2. Какие отношения вы считаете наиболее привлекательными?")
	fmt.Fprintln(output, "   Варианты: гетеро, би, гомо, не уверен")
	fmt.Fprintln(output, "3. Как часто вы испытывали чувства к людям своего пола?")
	fmt.Fprintln(output, "   Варианты: никогда, иногда, часто, постоянно")
	fmt.Fprintln(output, "4. Как часто вы испытывали чувства к людям противоположного пола?")
	fmt.Fprintln(output, "   Варианты: постоянно, часто, иногда, никогда")
	fmt.Fprintln(output, "5. Считаете ли вы, что ориентация может меняться?")
	fmt.Fprintln(output, "   Варианты: нет, редко, да, не уверен")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Для прохождения полного теста запустите программу без параметров.")
	fmt.Fprintln(output, "Пример: rb10.exe")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Результат демо-теста (усреднённый):")
	fmt.Fprintln(output, "Скорее всего, ваша ориентация уникальна и не укладывается в простые категории.")
	fmt.Fprintln(output, "Помните: ориентация – это сложный и индивидуальный аспект личности.")
	fmt.Fprintln(output, "Если вы испытываете дискомфорт, обратитесь к психологу или в поддерживающие организации.")
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
	case *ArrayAssignmentStatement:
		fmt.Fprintf(output, "%sArrayAssignment: %s[", prefix, n.Name)
		printAST(n.Index, 0)
		fmt.Fprint(output, "] = \n")
		printAST(n.Value, indent+1)
	case *PrintStatement:
		fmt.Fprintln(output, prefix+"Print")
		printAST(n.Value, indent+1)
	case *IfStatement:
		fmt.Fprintln(output, prefix+"If")
		fmt.Fprintln(output, prefix+"  Condition:")
		printAST(n.Condition, indent+2)
		fmt.Fprintln(output, prefix+"  Then:")
		for _, stmt := range n.ThenBlock {
			printAST(stmt, indent+2)
		}
		if len(n.ElseIfBlocks) > 0 {
			fmt.Fprintln(output, prefix+"  ElseIf:")
			for _, elseIf := range n.ElseIfBlocks {
				fmt.Fprintln(output, prefix+"    Condition:")
				printAST(elseIf.Condition, indent+3)
				fmt.Fprintln(output, prefix+"    Block:")
				for _, stmt := range elseIf.Block {
					printAST(stmt, indent+3)
				}
			}
		}
		if len(n.ElseBlock) > 0 {
			fmt.Fprintln(output, prefix+"  Else:")
			for _, stmt := range n.ElseBlock {
				printAST(stmt, indent+2)
			}
		}
	case *WhileStatement:
		fmt.Fprintln(output, prefix+"While")
		fmt.Fprintln(output, prefix+"  Condition:")
		printAST(n.Condition, indent+2)
		fmt.Fprintln(output, prefix+"  Body:")
		for _, stmt := range n.Body {
			printAST(stmt, indent+2)
		}
	case *BinaryOpNode:
		fmt.Fprintf(output, "%sBinaryOp: %s\n", prefix, n.Op)
		fmt.Fprintln(output, prefix+"  Left:")
		printAST(n.Left, indent+2)
		fmt.Fprintln(output, prefix+"  Right:")
		printAST(n.Right, indent+2)
	case *UnaryNode:
		fmt.Fprintf(output, "%sUnary: %s\n", prefix, n.Op)
		printAST(n.Expr, indent+1)
	case *NumberNode:
		fmt.Fprintf(output, "%sNumber: %d\n", prefix, n.Value)
	case *FloatNode:
		fmt.Fprintf(output, "%sFloat: %f\n", prefix, n.Value)
	case *StringNode:
		fmt.Fprintf(output, "%sString: \"%s\"\n", prefix, n.Value)
	case *BooleanNode:
		fmt.Fprintf(output, "%sBoolean: %t\n", prefix, n.Value)
	case *VariableNode:
		fmt.Fprintf(output, "%sVariable: %s\n", prefix, n.Name)
	case *ArrayNode:
		fmt.Fprintln(output, prefix+"Array")
		for _, elem := range n.Elements {
			printAST(elem, indent+1)
		}
	case *ArrayIndexNode:
		fmt.Fprintf(output, "%sArrayIndex: %s[", prefix, n.Name)
		printAST(n.Index, 0)
		fmt.Fprintln(output, "]")
	case *HelpStatement:
		fmt.Fprintln(output, prefix+"Help")
		printAST(n.Country, indent+1)
	case *OrientationStatement:
		fmt.Fprintln(output, prefix+"Orientation")
	case *FunctionDeclaration:
		exported := ""
		if n.Exported {
			exported = " (exported)"
		}
		fmt.Fprintf(output, "%sFunctionDeclaration%s: %s(%s)\n", prefix, exported, n.Name, strings.Join(n.Params, ", "))
		for _, stmt := range n.Body {
			printAST(stmt, indent+1)
		}
	case *FunctionCall:
		fmt.Fprintf(output, "%sFunctionCall: %s\n", prefix, n.Name)
		for _, arg := range n.Args {
			printAST(arg, indent+1)
		}
	case *ReturnStatement:
		fmt.Fprintln(output, prefix+"Return")
		printAST(n.Value, indent+1)
	case *TryCatchStatement:
		fmt.Fprintln(output, prefix+"TryCatch")
		fmt.Fprintln(output, prefix+"  Try:")
		for _, stmt := range n.TryBlock {
			printAST(stmt, indent+2)
		}
		fmt.Fprintln(output, prefix+"  Catch:")
		for _, stmt := range n.CatchBlock {
			printAST(stmt, indent+2)
		}
	case *ExpressionStatement:
		fmt.Fprintln(output, prefix+"ExpressionStatement")
		printAST(n.Expr, indent+1)
	default:
		fmt.Fprintf(output, "%sUnknown node: %T\n", prefix, n)
	}
}

// ---------- runExample ----------
func runExample() {
	program := `
		@ Пример программы с новыми социальными функциями
		
		RAINBOW main() {
			COMINGOUT "🌈 Добро пожаловать в LGBTScript с поддержкой сообщества!";
			
			@ Проверка ресурсов
			LESBIAN support = getCrisisSupport("Россия", "горячая_линия");
			COMINGOUT support;
			
			@ Ежедневная аффирмация
			LESBIAN affirmation = getDailyAffirmation("self-love");
			COMINGOUT affirmation;
			
			@ Поиск безопасных мест
			LESBIAN spaces = findSafeSpace("кафе", "Москва", 5);
			COMINGOUT spaces;
			
			@ Проверка эмоционального состояния
			GENDER moods = ["тревога", "одиночество"];
			LESBIAN moodResult = moodCheck(moods, true);
			COMINGOUT moodResult;
			
			@ Определение термина
			LESBIAN term = defineTerm("небинарность", "ru");
			COMINGOUT term;
			
			@ Рекомендация книги
			LESBIAN book = getLGBTQBook("фантастика");
			COMINGOUT book;
			
			@ Плейлист
			LESBIAN playlist = getLGBTQPlaylist("empowerment");
			COMINGOUT playlist;
			
			@ Мероприятия
			LESBIAN events = getLGBTQEvents(30, "online");
			COMINGOUT events;
			
			RETURN 0;
		}
		
		MAIN();
	`

	fmt.Fprintln(output, "=== LGBTScript с социальными функциями ===")
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

// ---------- Главная функция ----------
func main() {
	setupConsole()

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
	lgbtFile := flag.String("lgbt", "", "исполнить файл с кодом (аналог позиционного аргумента)")
	debug := flag.Bool("debug", false, "включить режим отладки")
	example := flag.Bool("example", false, "показать расширенный пример с социальными функциями")
	flag.Parse()

	debugMode = *debug

	if *example {
		runExample()
		return
	}

	if *lgbtFile != "" {
		isFileExecution = true
		currentFilePath = *lgbtFile
	}

	if *lgbtFile != "" {
		data, err := os.ReadFile(*lgbtFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения файла %s: %v\n", *lgbtFile, err)
			os.Exit(1)
		}
		program := string(data)

		lexer := NewLexer(program)
		tokens, err := lexer.Tokenize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Лексическая ошибка: %v\n", err)
			os.Exit(1)
		}
		if *showTokens {
			fmt.Fprintln(output, "=== Токены ===")
			for _, t := range tokens {
				fmt.Fprintf(output, "%v\n", t)
			}
			fmt.Fprintln(output)
		}

		parser := NewParser(tokens)
		ast, err := parser.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Синтаксическая ошибка: %v\n", err)
			os.Exit(1)
		}
		if *showAST {
			fmt.Fprintln(output, "=== AST ===")
			printAST(ast, 0)
			fmt.Fprintln(output)
		}

		interpreter := NewInterpreter()
		_, err = interpreter.EvaluateProgram(ast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *command != "" {
		interpreter := NewInterpreter()
		err := executeCode(*command, interpreter)
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
		program := string(data)

		lexer := NewLexer(program)
		tokens, err := lexer.Tokenize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Лексическая ошибка: %v\n", err)
			os.Exit(1)
		}
		if *showTokens {
			fmt.Fprintln(output, "=== Токены ===")
			for _, t := range tokens {
				fmt.Fprintf(output, "%v\n", t)
			}
			fmt.Fprintln(output)
		}

		parser := NewParser(tokens)
		ast, err := parser.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Синтаксическая ошибка: %v\n", err)
			os.Exit(1)
		}
		if *showAST {
			fmt.Fprintln(output, "=== AST ===")
			printAST(ast, 0)
			fmt.Fprintln(output)
		}

		interpreter := NewInterpreter()
		_, err = interpreter.EvaluateProgram(ast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Если нет аргументов - показываем демо
	fmt.Fprintln(output, "🌈 LGBTScript - Язык программирования с поддержкой ЛГБТ+ сообщества")
	fmt.Fprintln(output, "📖 Используйте --example для демонстрации социальных функций")
	fmt.Fprintln(output, "📁 Укажите файл .rainbow для выполнения")
	runExample()
}