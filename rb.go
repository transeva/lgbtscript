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
			"lesbian": true, "gay": true, "trans": true, "nonbinary": true, "gender": true,
			"comingout": true, "cis": true, "nocis": true,
			"true": true, "false": true,
			"help": true, "orientation": true,
			"rainbow": true,
			"return": true,
			"try": true, "catch": true,
			"export": true,
			"asexual": true,
			"queer": true, "extends": true, "this": true, "super": true, "new": true,
			"pride": true, "homo": true,
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
			(ch == '&' && next == '&') || (ch == '|' && next == '|') {
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

type HomoStatement struct {
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

// ---------- ООП узлы с QUEER ----------
type QueerClassDeclaration struct {
	BaseNode
	Name       string
	Parent     string
	Fields     map[string]string
	Methods    map[string]*FunctionDeclaration
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
		case "lesbian", "gay", "trans", "nonbinary", "gender", "asexual":
			return p.parseTypedDeclaration()
		case "comingout":
			return p.parsePrintStatement()
		case "cis":
			return p.parseIfStatement()
		case "pride":
			return p.parseWhileStatement()
		case "homo":
			return p.parseHomoStatement()
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

func (p *Parser) parseQueerClassDeclaration() (Node, error) {
	token := p.peek()
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
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Name:     name.Value,
		Parent:   parent,
		Fields:   make(map[string]string),
		Methods:  make(map[string]*FunctionDeclaration),
	}
	
	for p.peek().Value != "}" && p.peek().Type != TOKEN_EOF {
		tok := p.peek()
		if tok.Type == TOKEN_KEYWORD {
			keyword := tok.Value
			switch keyword {
			case "lesbian", "gay", "trans", "nonbinary", "gender", "asexual":
				p.next()
				fieldName, err := p.expect(TOKEN_IDENTIFIER, "")
				if err != nil {
					return nil, err
				}
				class.Fields[fieldName.Value] = keyword
				if p.peek().Value == ";" {
					p.next()
				}
			case "rainbow":
				method, err := p.parseFunctionDeclaration()
				if err != nil {
					return nil, err
				}
				if fn, ok := method.(*FunctionDeclaration); ok {
					if fn.Name == "init" {
						class.Constructor = fn
					} else {
						class.Methods[fn.Name] = fn
					}
				}
			default:
				return nil, fmt.Errorf("unexpected keyword in class: %s", keyword)
			}
		} else {
			return nil, fmt.Errorf("expected field or method in class, got %s", tok.Value)
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
	
	return &QueerInstanceNode{
		BaseNode:  BaseNode{Line: token.Line, Col: token.Col},
		ClassName: className.Value,
		Args:      args,
	}, nil
}

func (p *Parser) parseFieldAccess(objName string) (Node, error) {
	token := p.peek()
	p.next()
	
	_, err := p.expect(TOKEN_OPERATOR, ".")
	if err != nil {
		return nil, err
	}
	
	field, err := p.expect(TOKEN_IDENTIFIER, "")
	if err != nil {
		return nil, err
	}
	
	if p.peek().Value == "(" {
		p.pos--
		methodName := field.Value
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
		
		return &QueerMethodCallNode{
			BaseNode: BaseNode{Line: token.Line, Col: token.Col},
			Object:   &VariableNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Name: objName},
			Method:   methodName,
			Args:     args,
		}, nil
	}
	
	return &QueerFieldAccessNode{
		BaseNode: BaseNode{Line: token.Line, Col: token.Col},
		Object:   &VariableNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Name: objName},
		Field:    field.Value,
	}, nil
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
	case "asexual":
		defaultValue = &NumberNode{BaseNode: BaseNode{Line: token.Line, Col: token.Col}, Value: 0}
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

func (p *Parser) parseHomoStatement() (Node, error) {
	token := p.peek()
	p.next()
	
	init, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	
	if p.peek().Value != ";" {
		return nil, fmt.Errorf("expected ';' after initialization in homo loop")
	}
	p.next()
	
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	if p.peek().Value != ";" {
		return nil, fmt.Errorf("expected ';' after condition in homo loop")
	}
	p.next()
	
	update, err := p.parseStatement()
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
	
	return &HomoStatement{
		BaseNode:  BaseNode{Line: token.Line, Col: token.Col},
		Init:      init,
		Condition: condition,
		Update:    update,
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
		if p.peekNext().Value == "." {
			return p.parseFieldAccess(token.Value)
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
		} else if keyword == "this" {
			p.next()
			return &ThisNode{BaseNode: BaseNode{Line: line, Col: col}}, nil
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

func NewTypedObject(obj *QueerInstance) TypedValue {
	return TypedValue{Type: TypeObject, Value: obj}
}

// ---------- QUEER классы во время выполнения ----------
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

// ---------- Интерпретатор ----------
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
		queerClasses:   make(map[string]*QueerClassDeclaration),
		instances:      make(map[string]*QueerInstance),
		exportedFuncs:  make(map[string]*FunctionDeclaration),
		callStack:      []callFrame{{vars: make(map[string]TypedValue), types: make(map[string]string)}},
		returnValue:    TypedValue{Type: TypeNull, Value: nil},
		returnFlag:     false,
		maxRecursion:   1000,
		recursionDepth: 0,
		sandbox:        NewSandbox(),
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		this:           nil,
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
	case "asexual":
		return TypeInt
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
	if handlerValue.Type != TypeObject {
		return TypedValue{}, fmt.Errorf("addRoute: fourth argument must be a function")
	}
	
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
// СОЦИАЛЬНЫЕ ФУНКЦИИ
// ============================================

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

// ============================================
// НОВЫЕ ЛГБТ-ФУНКЦИИ
// ============================================

func (i *Interpreter) getPrideParadeInfo(args []TypedValue) (TypedValue, error) {
	city := "global"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			city = c
		}
	}

	year := time.Now().Year()
	if len(args) > 1 {
		if y, ok := args[1].Value.(int); ok {
			year = y
		}
	}

	parades := map[string]string{
		"москва":   "🏳️‍🌈 Московский прайд: июнь 2026, ул. Тверская",
		"спб":      "🏳️‍🌈 Санкт-Петербургский прайд: июль 2026, Невский проспект",
		"нью-йорк": "🏳️‍🌈 NYC Pride: июнь 2026, Манхэттен",
		"лондон":   "🏳️‍🌈 London Pride: июль 2026, центр Лондона",
		"берлин":   "🏳️‍🌈 Berlin Pride (CSD): июль 2026, Бранденбургские ворота",
		"париж":    "🏳️‍🌈 Paris Pride: июнь 2026, Марсово поле",
		"сидней":   "🏳️‍🌈 Sydney Mardi Gras: февраль-март 2026, Оксфорд-стрит",
		"токио":    "🏳️‍🌈 Tokyo Rainbow Pride: апрель 2026, парк Ёёги",
		"сан-франциско": "🏳️‍🌈 SF Pride: июнь 2026, Маркет-стрит",
		"афины":    "🏳️‍🌈 Athens Pride: июнь 2026, площадь Синтагма",
	}

	result := fmt.Sprintf("🏳️‍🌈 Информация о парадах гордости (%s, %d год):\n", city, year)
	
	if info, ok := parades[strings.ToLower(city)]; ok {
		result += "  • " + info + "\n"
	} else {
		result += "  • ℹ️ Информация о параде в этом городе уточняется\n"
	}
	
	result += "\n📅 Рекомендуем проверить даты на официальных сайтах прайд-организаций."
	return NewTypedString(result), nil
}

func (i *Interpreter) getComingOutTips(args []TypedValue) (TypedValue, error) {
	audience := "родители"
	if len(args) > 0 {
		if a, ok := args[0].Value.(string); ok {
			audience = a
		}
	}

	tips := map[string]string{
		"родители": "👨‍👩‍👦 Советы для каминг-аута перед родителями:\n" +
			"  • Выберите спокойное время для разговора\n" +
			"  • Будьте готовы к вопросам и эмоциям\n" +
			"  • Напомните им, что вы все еще их ребенок\n" +
			"  • Дайте им время на осмысление",
		"друзья": "👫 Советы для каминг-аута перед друзьями:\n" +
			"  • Начните с самых близких друзей\n" +
			"  • Будьте честны и открыты\n" +
			"  • Дайте им понять, что вы доверяете им",
		"работа": "💼 Советы для каминг-аута на работе:\n" +
			"  • Проверьте политику компании по ЛГБТ+\n" +
			"  • Поговорите с HR или доверенным менеджером\n" +
			"  • Оцените риски в вашей стране",
		"школа": "🏫 Советы для каминг-аута в школе:\n" +
			"  • Найдите поддерживающего учителя или психолога\n" +
			"  • Убедитесь, что школа поддерживает ЛГБТ+\n" +
			"  • Не торопитесь, делайте это в своем темпе",
	}

	result := fmt.Sprintf("💡 Советы по каминг-ауту (аудитория: %s):\n", audience)
	
	if tipsContent, ok := tips[strings.ToLower(audience)]; ok {
		result += tipsContent + "\n"
	} else {
		result += "  • ℹ️ Общие советы по каминг-ауту:\n"
		result += "  • Будьте собой и доверяйте своим чувствам\n"
		result += "  • Ищите поддержку в ЛГБТ+ сообществе\n"
	}
	
	result += "\n🔗 Дополнительная поддержка: https://comingout.org"

	return NewTypedString(result), nil
}

func (i *Interpreter) getTransHealthcare(args []TypedValue) (TypedValue, error) {
	country := "россия"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	healthcareInfo := map[string]string{
		"россия": "🏳️‍⚧️ Транс-здравоохранение в России:\n" +
			"  • Гендерно-аффирмативная помощь ограничена\n" +
			"  • Требуется психиатрическое заключение\n" +
			"  • Гормональная терапия доступна после диагноза\n" +
			"  • Ресурсы: Транс-Альянс, Транс-Помощь",
		"сша": "🏳️‍⚧️ Транс-здравоохранение в США:\n" +
			"  • Широкий доступ к гендерно-аффирмативной помощи\n" +
			"  • Медицинское страхование часто покрывает лечение\n" +
			"  • Доступны клиники, специализирующиеся на транс-здоровье",
		"великобритания": "🏳️‍⚧️ Транс-здравоохранение в Великобритании:\n" +
			"  • NHS предоставляет гендерно-аффирмативную помощь\n" +
			"  • Возможны длительные очереди\n" +
			"  • Существуют частные клиники",
		"германия": "🏳️‍⚧️ Транс-здравоохранение в Германии:\n" +
			"  • Страховка покрывает гендерно-аффирмативную помощь\n" +
			"  • Доступны специализированные центры\n" +
			"  • Требуется психиатрическое заключение",
	}

	result := "🏥 Информация о транс-здравоохранении:\n"
	if info, ok := healthcareInfo[strings.ToLower(country)]; ok {
		result += info + "\n"
	} else {
		result += "  • ℹ️ Информация для этой страны уточняется\n"
	}
	
	result += "\n💡 Рекомендуем обратиться к местным транс-организациям."

	return NewTypedString(result), nil
}

func (i *Interpreter) findLGBTQShelter(args []TypedValue) (TypedValue, error) {
	location := "москва"
	if len(args) > 0 {
		if l, ok := args[0].Value.(string); ok {
			location = l
		}
	}

	shelters := []string{
		"🏠 ЛГБТ-убежище 'Свет' - круглосуточная поддержка",
		"🏠 Приют для ЛГБТ+ молодежи 'Надежда'",
		"🏠 Центр временного проживания 'Радужный дом'",
		"🏠 Кризисный центр для ЛГБТ 'Вместе'",
	}

	result := fmt.Sprintf("🏠 ЛГБТ-убежища и приюты в регионе '%s':\n", location)
	for _, shelter := range shelters {
		result += "  • " + shelter + "\n"
	}
	result += "\n📞 Телефон кризисной поддержки: 8-800-XXX-XX-XX"

	return NewTypedString(result), nil
}

func (i *Interpreter) getIntersexResources(args []TypedValue) (TypedValue, error) {
	country := "global"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	resources := "🌍 Ресурсы для интерсекс-людей:\n" +
		"  • OII (Organization Intersex International)\n" +
		"  • InterACT - защита прав интерсекс-молодежи\n" +
		"  • Ресурсы по медицинской помощи\n" +
		"  • Группы поддержки для интерсекс-людей\n"

	if country != "global" {
		resources += fmt.Sprintf("\n📍 Ресурсы в стране '%s' уточняются.\n", country)
	}

	return NewTypedString(resources), nil
}

func (i *Interpreter) getNonbinaryGuide(args []TypedValue) (TypedValue, error) {
	guide := "💜 Гид для небинарных людей:\n\n" +
		"📝 Идентичность:\n" +
		"  • Небинарность - гендерная идентичность вне бинарной системы\n" +
		"  • Может включать: агендерность, бигендерность, гендерфлюидность\n" +
		"  • У каждого свой уникальный опыт\n\n" +
		"🔤 Язык и местоимения:\n" +
		"  • Местоимения: они/их, ze/zir, или другие\n" +
		"  • Важно уважать выбор человека\n\n" +
		"💡 Советы:\n" +
		"  • Найдите поддерживающее сообщество\n" +
		"  • Практикуйте самовыражение в безопасной среде\n" +
		"  • Помните, что ваша идентичность валидна"

	return NewTypedString(guide), nil
}

func (i *Interpreter) findLGBTQTherapist(args []TypedValue) (TypedValue, error) {
	specialty := "психотерапевт"
	if len(args) > 0 {
		if s, ok := args[0].Value.(string); ok {
			specialty = s
		}
	}

	location := "онлайн"
	if len(args) > 1 {
		if l, ok := args[1].Value.(string); ok {
			location = l
		}
	}

	therapists := []string{
		"🧠 Специалист по ЛГБТ+ вопросам - онлайн-консультации",
		"🧠 Терапевт с опытом работы с транс-людьми",
		"🧠 Психолог для небинарных и гендерно-неконформных людей",
		"🧠 Семейный терапевт для ЛГБТ+ семей",
		"🧠 Кризисный психолог - поддержка при каминг-ауте",
	}

	result := fmt.Sprintf("🧠 ЛГБТ-дружественные терапевты (%s, %s):\n", specialty, location)
	for _, therapist := range therapists {
		result += "  • " + therapist + "\n"
	}
	result += "\n💡 Проверьте лицензию и отзывы перед обращением"

	return NewTypedString(result), nil
}

func (i *Interpreter) getAsylumInfo(args []TypedValue) (TypedValue, error) {
	country := "россия"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	info := "🛂 Информация о получении убежища для ЛГБТ+:\n\n" +
		"📋 Требования:\n" +
		"  • Доказательства преследования по признаку ориентации\n" +
		"  • Медицинские документы (при необходимости)\n" +
		"  • Письма поддержки от ЛГБТ-организаций\n\n" +
		"📍 Страны для убежища:\n" +
		"  • Канада - программа защиты ЛГБТ-беженцев\n" +
		"  • Великобритания - специальные программы\n" +
		"  • Германия - поддержка квир-беженцев\n" +
		"  • США - убежище для преследуемых ЛГБТ+\n"

	if country != "россия" {
		info += fmt.Sprintf("\n📍 Информация для страны '%s' уточняется.\n", country)
	}

	return NewTypedString(info), nil
}

func (i *Interpreter) getAsexualResources(args []TypedValue) (TypedValue, error) {
	resources := "🤍 Ресурсы для асексуальных людей:\n\n" +
		"📖 Определение:\n" +
		"  • Асексуальность - отсутствие сексуального влечения\n" +
		"  • Спектр: демисексуальность, грей-сексуальность\n" +
		"  • Аромантизм - отсутствие романтического влечения\n\n" +
		"💡 Сообщество:\n" +
		"  • AVEN (Asexual Visibility and Education Network)\n" +
		"  • Группы поддержки для асексуалов\n" +
		"  • Онлайн-форумы и чаты\n\n" +
		"📚 Ресурсы:\n" +
		"  • Книги и статьи об асексуальности\n" +
		"  • Документальные фильмы\n" +
		"  • Подкасты"

	return NewTypedString(resources), nil
}

func (i *Interpreter) getPolyamoryGuide(args []TypedValue) (TypedValue, error) {
	guide := "💕 Гид по полиамории и этичной немоногамии:\n\n" +
		"📖 Основные концепции:\n" +
		"  • Этичная немоногамия - отношения с согласием всех участников\n" +
		"  • Полиамория - возможность любить нескольких людей\n" +
		"  • Различные модели: иерархическая, неиерархическая\n\n" +
		"⚖️ Правила и границы:\n" +
		"  • Открытая коммуникация\n" +
		"  • Честность и прозрачность\n" +
		"  • Уважение к потребностям каждого\n\n" +
		"💡 Ресурсы:\n" +
		"  • Книга 'Этичный шлюха'\n" +
		"  • Сообщества и группы поддержки\n" +
		"  • Психологи, специализирующиеся на немоногамии"

	return NewTypedString(guide), nil
}

func (i *Interpreter) getGenderAffirmingCare(args []TypedValue) (TypedValue, error) {
	country := "россия"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			country = c
		}
	}

	info := "🏳️‍⚧️ Гендерно-аффирмативная помощь:\n\n" +
		"🩺 Медицинские услуги:\n" +
		"  • Гормональная терапия\n" +
		"  • Хирургические операции\n" +
		"  • Логопедия для коррекции голоса\n" +
		"  • Электроэпиляция\n\n" +
		"🧠 Психологическая поддержка:\n" +
		"  • Терапия с ЛГБТ-дружественным психологом\n" +
		"  • Группы поддержки\n" +
		"  • Консультации по социальному переходу\n\n" +
		"⚖️ Юридическая помощь:\n" +
		"  • Смена документов\n" +
		"  • Защита прав на рабочем месте"

	if country != "россия" {
		info += fmt.Sprintf("\n📍 Информация для страны '%s' уточняется.\n", country)
	}

	return NewTypedString(info), nil
}

func (i *Interpreter) findLGBTQCommunity(args []TypedValue) (TypedValue, error) {
	interest := "общий"
	if len(args) > 0 {
		if i2, ok := args[0].Value.(string); ok {
			interest = i2
		}
	}

	communities := map[string]string{
		"спорт": "🏋️‍♀️ Спортивные ЛГБТ+ сообщества: клубы по футболу, волейболу, плаванию",
		"искусство": "🎨 Арт-сообщества: ЛГБТ+ художники, писатели, музыканты",
		"технологии": "💻 ЛГБТ+ в IT: сообщества разработчиков, дизайнеров",
		"образование": "📚 Образовательные сообщества: клубы, лекции, тренинги",
		"активизм": "✊ Активистские группы: защита прав, протесты, кампании",
		"здоровье": "🏥 ЛГБТ+ в здравоохранении: врачи, психологи, поддержка",
		"бизнес": "💼 ЛГБТ+ предприниматели: стартапы, нетворкинг",
		"религия": "🕊️ Инклюзивные религиозные общины",
		"родители": "👨‍👩‍👧 ЛГБТ-родители и семьи",
		"молодежь": "🧒 ЛГБТ+ молодежные организации",
	}

	result := fmt.Sprintf("👥 ЛГБТ+ сообщества по интересам: %s\n", interest)
	
	if communitiesContent, ok := communities[strings.ToLower(interest)]; ok {
		result += "  • " + communitiesContent + "\n"
	} else {
		result += "  • Общие ЛГБТ+ сообщества: встречи, группы поддержки, мероприятия\n"
	}
	
	result += "\n🔗 Поищите в социальных сетях и на платформах для ЛГБТ+"

	return NewTypedString(result), nil
}

func (i *Interpreter) getLGBTQHistory(args []TypedValue) (TypedValue, error) {
	era := "общая"
	if len(args) > 0 {
		if e, ok := args[0].Value.(string); ok {
			era = e
		}
	}

	history := map[string]string{
		"общая": "📜 Основные вехи ЛГБТ+ истории:\n" +
			"  • 1969 - Стоунволлские бунты\n" +
			"  • 1973 - Исключение гомосексуальности из DSM\n" +
			"  • 2015 - Легализация однополых браков в США\n" +
			"  • 2023 - Все больше стран легализуют однополые браки",
		"древний": "🏛️ Древняя история:\n" +
			"  • Древняя Греция - гомосексуальные отношения\n" +
			"  • Древний Рим - разнообразие сексуальных практик\n" +
			"  • Индия - историческое признание третьего пола",
		"средневековье": "⚔️ Средневековье:\n" +
			"  • Преследование за гомосексуальность\n" +
			"  • Подпольные ЛГБТ+ сообщества\n" +
			"  • Религиозные запреты",
		"новое": "🌍 Новейшая история:\n" +
			"  • ЛГБТ+ движение\n" +
			"  • Достижения в правах ЛГБТ+\n" +
			"  • Современные вызовы",
	}

	result := fmt.Sprintf("📚 ЛГБТ+ история (эпоха: %s):\n", era)
	
	if historyContent, ok := history[strings.ToLower(era)]; ok {
		result += historyContent + "\n"
	} else {
		result += "  • История ЛГБТ+ разнообразна и богата\n"
	}
	
	result += "\n📖 Рекомендуем книги и фильмы по ЛГБТ+ истории"

	return NewTypedString(result), nil
}

func (i *Interpreter) getLGBTQParenting(args []TypedValue) (TypedValue, error) {
	parenting := "👨‍👩‍👧 Информация для ЛГБТ-родителей:\n\n" +
		"📋 Пути создания семьи:\n" +
		"  • Суррогатное материнство\n" +
		"  • Усыновление\n" +
		"  • Донорство\n" +
		"  • Воспитание детей от предыдущих отношений\n\n" +
		"⚖️ Юридические вопросы:\n" +
		"  • Права обоих родителей\n" +
		"  • Усыновление в разных странах\n" +
		"  • Регистрация брака для защиты прав\n\n" +
		"💡 Поддержка:\n" +
		"  • Группы для ЛГБТ-родителей\n" +
		"  • Ресурсы для детей из ЛГБТ-семей\n" +
		"  • Консультации специалистов"

	return NewTypedString(parenting), nil
}

func (i *Interpreter) getConversionTherapyHelp(args []TypedValue) (TypedValue, error) {
	help := "🚫 Помощь жертвам конверсионной терапии:\n\n" +
		"📋 Что это:\n" +
		"  • Попытки изменить сексуальную ориентацию\n" +
		"  • Псевдонаучные методы\n" +
		"  • Запрещена во многих странах\n\n" +
		"💡 Что делать:\n" +
		"  • Обратитесь за помощью к профессионалам\n" +
		"  • Найдите поддержку в ЛГБТ+ сообществе\n" +
		"  • Подайте жалобу в соответствующие органы\n\n" +
		"📞 Горячие линии:\n" +
		"  • Линия поддержки жертв конверсионной терапии\n" +
		"  • Психологическая помощь\n" +
		"  • Юридическая консультация"

	return NewTypedString(help), nil
}

func (i *Interpreter) getLGBTQHousing(args []TypedValue) (TypedValue, error) {
	location := "москва"
	if len(args) > 0 {
		if l, ok := args[0].Value.(string); ok {
			location = l
		}
	}

	housing := fmt.Sprintf("🏠 ЛГБТ-дружественное жилье в %s:\n\n", location) +
		"🏢 Варианты:\n" +
		"  • ЛГБТ-дружественные общежития\n" +
		"  • Комнаты с квир-соседями\n" +
		"  • Квартиры в инклюзивных районах\n" +
		"  • Временные приюты\n\n" +
		"🔍 Как искать:\n" +
		"  • Специализированные платформы для ЛГБТ+\n" +
		"  • Группы в социальных сетях\n" +
		"  • Рекомендации от ЛГБТ-организаций\n\n" +
		"⚖️ Права:\n" +
		"  • Защита от дискриминации\n" +
		"  • Юридическая помощь"

	return NewTypedString(housing), nil
}

func (i *Interpreter) getQueerArt(args []TypedValue) (TypedValue, error) {
	medium := "все"
	if len(args) > 0 {
		if m, ok := args[0].Value.(string); ok {
			medium = m
		}
	}

	art := fmt.Sprintf("🎨 Квир-искусство (медиум: %s):\n\n", medium) +
		"🎭 Художники:\n" +
		"  • Фрида Кало - символика и идентичность\n" +
		"  • Кит Харинг - активистское искусство\n" +
		"  • Дэвид Хокни - квир-живопись\n" +
		"  • Zanele Muholi - фотография квир-сообществ\n\n" +
		"📚 Направления:\n" +
		"  • Квир-живопись и скульптура\n" +
		"  • Фотография и инсталляции\n" +
		"  • Перформанс и видео-арт\n" +
		"  • Квир-литература\n\n" +
		"🎬 Где смотреть:\n" +
		"  • ЛГБТ+ кинофестивали\n" +
		"  • Выставки в музеях\n" +
		"  • Онлайн-галереи"

	return NewTypedString(art), nil
}

func (i *Interpreter) getLGBTQFriendlyCities(args []TypedValue) (TypedValue, error) {
	criteria := "общий"
	if len(args) > 0 {
		if c, ok := args[0].Value.(string); ok {
			criteria = c
		}
	}

	cities := map[string]string{
		"общий": "🌍 ЛГБТ-дружественные города мира:\n" +
			"  • Сан-Франциско (США)\n" +
			"  • Амстердам (Нидерланды)\n" +
			"  • Берлин (Германия)\n" +
			"  • Торонто (Канада)\n" +
			"  • Сидней (Австралия)\n" +
			"  • Монреаль (Канада)\n" +
			"  • Копенгаген (Дания)",
		"европа": "🌍 ЛГБТ-дружественные города Европы:\n" +
			"  • Амстердам\n" +
			"  • Берлин\n" +
			"  • Копенгаген\n" +
			"  • Лондон\n" +
			"  • Мадрид",
		"азия": "🌍 ЛГБТ-дружественные города Азии:\n" +
			"  • Токио (Япония)\n" +
			"  • Сеул (Южная Корея)\n" +
			"  • Тайбэй (Тайвань)\n" +
			"  • Манила (Филиппины)",
		"америка": "🌍 ЛГБТ-дружественные города Америки:\n" +
			"  • Сан-Франциско\n" +
			"  • Торонто\n" +
			"  • Нью-Йорк\n" +
			"  • Монреаль\n" +
			"  • Буэнос-Айрес",
	}

	result := fmt.Sprintf("🌍 ЛГБТ-дружественные города (категория: %s):\n", criteria)
	
	if citiesContent, ok := cities[strings.ToLower(criteria)]; ok {
		result += citiesContent + "\n"
	} else {
		result += "  • Информация по этой категории уточняется\n"
	}
	
	result += "\n💡 Рейтинги основаны на законах, культуре и ЛГБТ-жизни"

	return NewTypedString(result), nil
}

// ---------- Встроенные функции ----------
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
		"findSafeSpace":               i.findSafeSpace,
		"getCrisisSupport":            i.getCrisisSupport,
		"getLGBTQLaws":                i.getLGBTQLaws,
		"getDailyAffirmation":         i.getDailyAffirmation,
		"moodCheck":                   i.moodCheck,
		"guidedBreathing":             i.guidedBreathing,
		"defineTerm":                  i.defineTerm,
		"lgbtHistoryQuiz":             i.lgbtHistoryQuiz,
		"getDailyFact":                i.getDailyFact,
		"getHRTInfo":                  i.getHRTInfo,
		"findLGBTDoctor":              i.findLGBTDoctor,
		"getDocumentChangeGuide":      i.getDocumentChangeGuide,
		"getLGBTQEvents":              i.getLGBTQEvents,
		"createLGBTQGroup":            i.createLGBTQGroup,
		"findVolunteerOpportunity":    i.findVolunteerOpportunity,
		"getLGBTQBook":                i.getLGBTQBook,
		"getLGBTQPlaylist":            i.getLGBTQPlaylist,
		"getLGBTQMovies":              i.getLGBTQMovies,
		"getPrideParadeInfo":          i.getPrideParadeInfo,
		"getComingOutTips":            i.getComingOutTips,
		"getTransHealthcare":          i.getTransHealthcare,
		"findLGBTQShelter":            i.findLGBTQShelter,
		"getIntersexResources":        i.getIntersexResources,
		"getNonbinaryGuide":           i.getNonbinaryGuide,
		"findLGBTQTherapist":          i.findLGBTQTherapist,
		"getAsylumInfo":               i.getAsylumInfo,
		"getAsexualResources":         i.getAsexualResources,
		"getPolyamoryGuide":           i.getPolyamoryGuide,
		"getGenderAffirmingCare":      i.getGenderAffirmingCare,
		"findLGBTQCommunity":          i.findLGBTQCommunity,
		"getLGBTQHistory":             i.getLGBTQHistory,
		"getLGBTQParenting":           i.getLGBTQParenting,
		"getConversionTherapyHelp":    i.getConversionTherapyHelp,
		"getLGBTQHousing":             i.getLGBTQHousing,
		"getQueerArt":                 i.getQueerArt,
		"getLGBTQFriendlyCities":      i.getLGBTQFriendlyCities,
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
	case TypeObject:
		return true
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
	case *HomoStatement:
		_, err := i.Evaluate(n.Init)
		if err != nil {
			return TypedValue{}, err
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
			
			_, err = i.Evaluate(n.Update)
			if err != nil {
				return TypedValue{}, err
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
	case *QueerClassDeclaration:
		i.mu.Lock()
		i.queerClasses[n.Name] = n
		i.mu.Unlock()
		return TypedValue{Type: TypeNull, Value: nil}, nil
	case *QueerInstanceNode:
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
			case "gay", "asexual":
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
					case "gay", "asexual":
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
				val, err := i.Evaluate(arg)
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
				_, err := i.Evaluate(stmt)
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
	case *QueerFieldAccessNode:
		objVal, err := i.Evaluate(n.Object)
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
	case *QueerMethodCallNode:
		objVal, err := i.Evaluate(n.Object)
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
			val, err := i.Evaluate(arg)
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
			lastResult, err = i.Evaluate(stmt)
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
		fmt.Fprintln(output, prefix+"While (pride)")
		fmt.Fprintln(output, prefix+"  Condition:")
		printAST(n.Condition, indent+2)
		fmt.Fprintln(output, prefix+"  Body:")
		for _, stmt := range n.Body {
			printAST(stmt, indent+2)
		}
	case *HomoStatement:
		fmt.Fprintln(output, prefix+"Homo (for)")
		fmt.Fprintln(output, prefix+"  Init:")
		printAST(n.Init, indent+2)
		fmt.Fprintln(output, prefix+"  Condition:")
		printAST(n.Condition, indent+2)
		fmt.Fprintln(output, prefix+"  Update:")
		printAST(n.Update, indent+2)
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
	case *QueerClassDeclaration:
		fmt.Fprintf(output, "%sQueerClass: %s", prefix, n.Name)
		if n.Parent != "" {
			fmt.Fprintf(output, " extends %s", n.Parent)
		}
		fmt.Fprintln(output)
		if len(n.Fields) > 0 {
			fmt.Fprintln(output, prefix+"  Fields:")
			for fieldName, fieldType := range n.Fields {
				fmt.Fprintf(output, "%s    %s: %s\n", prefix, fieldName, fieldType)
			}
		}
		if len(n.Methods) > 0 {
			fmt.Fprintln(output, prefix+"  Methods:")
			for methodName, method := range n.Methods {
				fmt.Fprintf(output, "%s    %s(%s)\n", prefix, methodName, strings.Join(method.Params, ", "))
			}
		}
		if n.Constructor != nil {
			fmt.Fprintf(output, "%s  Constructor: init(%s)\n", prefix, strings.Join(n.Constructor.Params, ", "))
		}
	case *QueerInstanceNode:
		fmt.Fprintf(output, "%sNew %s(", prefix, n.ClassName)
		for i, arg := range n.Args {
			if i > 0 {
				fmt.Fprint(output, ", ")
			}
			printAST(arg, 0)
		}
		fmt.Fprintln(output, ")")
	case *QueerFieldAccessNode:
		fmt.Fprintf(output, "%sQueerFieldAccess: ", prefix)
		printAST(n.Object, 0)
		fmt.Fprintf(output, ".%s", n.Field)
		fmt.Fprintln(output)
	case *QueerMethodCallNode:
		fmt.Fprintf(output, "%sQueerMethodCall: ", prefix)
		printAST(n.Object, 0)
		fmt.Fprintf(output, ".%s(", n.Method)
		for i, arg := range n.Args {
			if i > 0 {
				fmt.Fprint(output, ", ")
			}
			printAST(arg, 0)
		}
		fmt.Fprintln(output, ")")
	case *ThisNode:
		fmt.Fprintln(output, prefix+"this")
	case *SuperNode:
		fmt.Fprintln(output, prefix+"super")
	default:
		fmt.Fprintf(output, "%sUnknown node: %T\n", prefix, n)
	}
}

// ---------- runExample ----------
func runExample() {
	program := `
		@ Программа с циклами pride и homo, QUEER классами, социальными и серверными функциями
		
		QUEER Person {
			LESBIAN name;
			GAY age;
			LESBIAN gender;
			
			RAINBOW init(nameVal, ageVal, genderVal) {
				this.name = nameVal;
				this.age = ageVal;
				this.gender = genderVal;
				COMINGOUT "🔨 Создан Person: " + this.name;
			}
			
			RAINBOW introduce() {
				COMINGOUT "👋 Привет! Я " + this.name + ", мне " + this.age + " лет";
			}
		}
		
		QUEER Student EXTENDS Person {
			LESBIAN university;
			
			RAINBOW init(nameVal, ageVal, genderVal, universityVal) {
				SUPER.init(nameVal, ageVal, genderVal);
				this.university = universityVal;
			}
			
			RAINBOW study() {
				COMINGOUT "📚 " + this.name + " учится в " + this.university;
			}
		}
		
		RAINBOW main() {
			COMINGOUT "🌈 LGBTScript с циклами pride и homo, QUEER классами, социальными и серверными функциями";
			COMINGOUT "";
			
			COMINGOUT "📊 Цикл pride (while):";
			GAY i = 0;
			PRIDE (i < 5) {
				COMINGOUT "  pride #" + i;
				i = i + 1;
			}
			
			COMINGOUT "";
			COMINGOUT "📊 Цикл homo (for):";
			HOMO (GAY j = 0; j < 3; j = j + 1) {
				COMINGOUT "  homo #" + j;
			}
			
			COMINGOUT "";
			COMINGOUT "📊 Вложенные циклы:";
			HOMO (GAY a = 0; a < 3; a = a + 1) {
				PRIDE (GAY b = 0; b < 2; b = b + 1) {
					COMINGOUT "  a=" + a + ", b=" + b;
				}
			}
			
			COMINGOUT "";
			COMINGOUT "📊 Сумма чисел от 1 до 10:";
			GAY sum = 0;
			HOMO (GAY n = 1; n <= 10; n = n + 1) {
				sum = sum + n;
			}
			COMINGOUT "  Сумма = " + sum;
			
			COMINGOUT "";
			GENDER person = NEW Person("Алекс", 25, "мужской");
			person.introduce();
			
			GENDER student = NEW Student("Мария", 20, "женский", "МГУ");
			student.introduce();
			student.study();
			
			COMINGOUT "";
			COMINGOUT "🏳️‍🌈 Информация о парадах гордости:";
			COMINGOUT getPrideParadeInfo("москва", 2026);
			
			COMINGOUT "";
			COMINGOUT "💡 Советы по каминг-ауту:";
			COMINGOUT getComingOutTips("родители");
			
			COMINGOUT "";
			COMINGOUT "📊 Создание и управление серверами:";
			COMINGOUT createServer("myapi", 8080);
			COMINGOUT getServerStatus("myapi");
			
			COMINGOUT "📋 Список серверов:";
			COMINGOUT listServers();
			
			COMINGOUT "";
			COMINGOUT "🌍 ЛГБТ-дружественные города:";
			COMINGOUT getLGBTQFriendlyCities("европа");
			
			COMINGOUT "";
			COMINGOUT "💖 Любовь побеждает всё!";
			RETURN 0;
		}
		
		main();
	`

	fmt.Fprintln(output, "=== LGBTScript с циклами pride и homo, QUEER классами ===")
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
	example := flag.Bool("example", false, "показать расширенный пример с циклами, QUEER классами, социальными и серверными функциями")
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

	fmt.Fprintln(output, "🌈 LGBTScript - Язык программирования с QUEER классами, циклами и поддержкой ЛГБТ+ сообщества")
	fmt.Fprintln(output, "📖 Используйте --example для демонстрации всех возможностей")
	fmt.Fprintln(output, "📁 Укажите файл .rainbow для выполнения")
	runExample()
}