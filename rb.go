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
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
			"lesbian": true, "gay": true, "trans": true, "nonbinary": true,
			"comingout": true, "gender": true, "queer": true, "pride": true,
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
	if l.keywords[value] {
		l.tokens = append(l.tokens, Token{TOKEN_KEYWORD, value, startLine, startCol})
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
type Node interface{}

type NumberNode struct {
	Value int
}
type FloatNode struct {
	Value float64
}
type StringNode struct {
	Value string
}
type BooleanNode struct {
	Value bool
}
type VariableNode struct {
	Name string
}
type ArrayNode struct {
	Elements []Node
}
type ArrayIndexNode struct {
	Name  string
	Index Node
}

type BinaryOpNode struct {
	Left  Node
	Op    string
	Right Node
}

type TypedDeclaration struct {
	Type  string
	Name  string
	Value Node
}

type AssignmentStatement struct {
	Name  string
	Value Node
}

type ArrayAssignmentStatement struct {
	Name  string
	Index Node
	Value Node
}

type PrintStatement struct {
	Value Node
}

type IfStatement struct {
	Condition Node
	ThenBlock []Node
	ElseBlock []Node
}

type WhileStatement struct {
	Condition Node
	Body      []Node
}

type HelpStatement struct {
	Country Node
}

type OrientationStatement struct {
}

type FunctionDeclaration struct {
	Name     string
	Params   []string
	Body     []Node
	Exported bool
}

type FunctionCall struct {
	Name string
	Args []Node
}

type ReturnStatement struct {
	Value Node
}

type CommentStatement struct {
	Text string
}

type IncludeStatement struct {
	Filename string
}

type TryCatchStatement struct {
	TryBlock   []Node
	CatchBlock []Node
}

type ExpressionStatement struct {
	Expr Node
}

type Program struct {
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

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}
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

	if token.Type == TOKEN_INCLUDE {
		filename := token.Value
		p.next()
		return &IncludeStatement{Filename: filename}, nil
	}

	if token.Type == TOKEN_COMMENT {
		comment := &CommentStatement{Text: token.Value}
		p.next()
		return comment, nil
	}

	if token.Type == TOKEN_KEYWORD {
		switch token.Value {
		case "lesbian", "gay", "trans", "nonbinary":
			return p.parseTypedDeclaration()
		case "comingout":
			return p.parsePrintStatement()
		case "gender":
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
		return &ExpressionStatement{Expr: expr}, nil
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &ExpressionStatement{Expr: expr}, nil
}

func (p *Parser) parseExportStatement() (Node, error) {
	p.next()
	token := p.peek()
	if token.Type == TOKEN_KEYWORD && token.Value == "rainbow" {
		fn, err := p.parseFunctionDeclaration()
		if err != nil {
			return nil, err
		}
		if fnDecl, ok := fn.(*FunctionDeclaration); ok {
			fnDecl.Exported = true
		}
		return fn, nil
	}
	return nil, fmt.Errorf("expected function declaration after export")
}

func (p *Parser) parseTryCatchStatement() (Node, error) {
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
		TryBlock:   tryBlock,
		CatchBlock: catchBlock,
	}, nil
}

func (p *Parser) parseArrayAssignment(name string) (Node, error) {
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
	return &ArrayAssignmentStatement{Name: name, Index: index, Value: value}, nil
}

func (p *Parser) parseReturnStatement() (Node, error) {
	p.next()
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &ReturnStatement{Value: value}, nil
}

func (p *Parser) parseFunctionDeclaration() (Node, error) {
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
		Name:   name,
		Params: params,
		Body:   body,
	}, nil
}

func (p *Parser) parseFunctionCall(name string) (Node, error) {
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

	return &FunctionCall{Name: name, Args: args}, nil
}

func (p *Parser) parseTypedDeclaration() (Node, error) {
	typeToken := p.next()
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
			Type:  typeToken.Value,
			Name:  name.Value,
			Value: value,
		}, nil
	}

	_, err = p.expect(TOKEN_OPERATOR, ";")
	if err != nil {
		return nil, err
	}

	// Значения по умолчанию для каждого типа
	var defaultValue Node
	switch typeToken.Value {
	case "lesbian":
		defaultValue = &StringNode{Value: ""} // Строка по умолчанию - пустая строка
	case "gay":
		defaultValue = &NumberNode{Value: 0} // Число по умолчанию - 0
	case "trans":
		defaultValue = &FloatNode{Value: 0.0} // Float по умолчанию - 0.0
	case "nonbinary":
		defaultValue = &BooleanNode{Value: false} // Boolean по умолчанию - false
	default:
		defaultValue = &NumberNode{Value: 0}
	}

	return &TypedDeclaration{
		Type:  typeToken.Value,
		Name:  name.Value,
		Value: defaultValue,
	}, nil
}

func (p *Parser) parseAssignment() (Node, error) {
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
	return &AssignmentStatement{Name: name.Value, Value: value}, nil
}

func (p *Parser) parsePrintStatement() (Node, error) {
	p.next()
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &PrintStatement{Value: value}, nil
}

func (p *Parser) parseIfStatement() (Node, error) {
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

	var elseBlock []Node
	if p.peek().Type == TOKEN_KEYWORD && p.peek().Value == "queer" {
		p.next()
		_, err := p.expect(TOKEN_OPERATOR, "{")
		if err != nil {
			return nil, err
		}
		elseBlock, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TOKEN_OPERATOR, "}")
		if err != nil {
			return nil, err
		}
	}

	return &IfStatement{
		Condition: condition,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}, nil
}

func (p *Parser) parseWhileStatement() (Node, error) {
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
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *Parser) parseHelpStatement() (Node, error) {
	p.next()
	country, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Value == ";" {
		p.next()
	}
	return &HelpStatement{Country: country}, nil
}

func (p *Parser) parseOrientationStatement() (Node, error) {
	p.next()
	if p.peek().Value == ";" {
		p.next()
	}
	return &OrientationStatement{}, nil
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
		left = &BinaryOpNode{Left: left, Op: op, Right: right}
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
		left = &BinaryOpNode{Left: left, Op: op, Right: right}
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
		left = &BinaryOpNode{Left: left, Op: op, Right: right}
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
		left = &BinaryOpNode{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMultiplicative() (Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().Value == "*" || p.peek().Value == "/" || p.peek().Value == "%" {
		op := p.next().Value
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	token := p.peek()

	switch token.Type {
	case TOKEN_NUMBER:
		p.next()
		val, _ := strconv.Atoi(token.Value)
		return &NumberNode{Value: val}, nil
	case TOKEN_FLOAT:
		p.next()
		val, _ := strconv.ParseFloat(token.Value, 64)
		return &FloatNode{Value: val}, nil
	case TOKEN_STRING:
		p.next()
		return &StringNode{Value: token.Value}, nil
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
		return &ArrayNode{Elements: elements}, nil
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
			return &ArrayIndexNode{Name: name, Index: index}, nil
		}
		if p.peekNext().Value == "(" {
			return p.parseFunctionCall(token.Value)
		}
		p.next()
		return &VariableNode{Name: token.Value}, nil
	case TOKEN_KEYWORD:
		if token.Value == "true" {
			p.next()
			return &BooleanNode{Value: true}, nil
		} else if token.Value == "false" {
			p.next()
			return &BooleanNode{Value: false}, nil
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

// ---------- Интерпретатор ----------
type Interpreter struct {
	variables     map[string]interface{}
	variableTypes map[string]string
	functions     map[string]*FunctionDeclaration
	exportedFuncs map[string]*FunctionDeclaration
	callStack     []callFrame
	returnValue   interface{}
	returnFlag    bool
	errorHandler  func(error)
}

type callFrame struct {
	vars  map[string]interface{}
	types map[string]string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:     make(map[string]interface{}),
		variableTypes: make(map[string]string),
		functions:     make(map[string]*FunctionDeclaration),
		exportedFuncs: make(map[string]*FunctionDeclaration),
		callStack:     []callFrame{{vars: make(map[string]interface{}), types: make(map[string]string)}},
		returnValue:   nil,
		returnFlag:    false,
	}
}

func (i *Interpreter) pushFrame() {
	i.callStack = append(i.callStack, callFrame{
		vars:  make(map[string]interface{}),
		types: make(map[string]string),
	})
}

func (i *Interpreter) popFrame() {
	if len(i.callStack) > 1 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
}

func (i *Interpreter) getVar(name string) (interface{}, bool) {
	for idx := len(i.callStack) - 1; idx >= 0; idx-- {
		if val, ok := i.callStack[idx].vars[name]; ok {
			return val, true
		}
	}
	return nil, false
}

func (i *Interpreter) setVar(name string, value interface{}) {
	top := len(i.callStack) - 1
	i.callStack[top].vars[name] = value
}

func (i *Interpreter) getType(name string) (string, bool) {
	for idx := len(i.callStack) - 1; idx >= 0; idx-- {
		if typ, ok := i.callStack[idx].types[name]; ok {
			return typ, true
		}
	}
	return "", false
}

func (i *Interpreter) setType(name string, typ string) {
	top := len(i.callStack) - 1
	i.callStack[top].types[name] = typ
}

// ---------- Встроенные функции ----------
func (i *Interpreter) getBuiltinFunction(name string) (func([]interface{}) (interface{}, error), bool) {
	builtins := map[string]func([]interface{}) (interface{}, error){
		"readFile": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("readFile: expected 1 argument, got %d", len(args))
			}
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("readFile: argument must be string")
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			return string(data), nil
		},
		"writeFile": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("writeFile: expected 2 arguments, got %d", len(args))
			}
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("writeFile: first argument must be string")
			}
			content, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("writeFile: second argument must be string")
			}
			err := os.WriteFile(filename, []byte(content), 0644)
			return nil, err
		},
		"fileExists": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("fileExists: expected 1 argument, got %d", len(args))
			}
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("fileExists: argument must be string")
			}
			_, err := os.Stat(filename)
			return err == nil, nil
		},
		"getDirFiles": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("getDirFiles: expected 1 argument, got %d", len(args))
			}
			dir, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("getDirFiles: argument must be string")
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, err
			}
			var files []interface{}
			for _, entry := range entries {
				files = append(files, entry.Name())
			}
			return files, nil
		},
		"split": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("split: expected 2 arguments, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("split: first argument must be string")
			}
			delim, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("split: second argument must be string")
			}
			parts := strings.Split(text, delim)
			result := make([]interface{}, len(parts))
			for i, p := range parts {
				result[i] = p
			}
			return result, nil
		},
		"replace": func(args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("replace: expected 3 arguments, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("replace: first argument must be string")
			}
			old, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("replace: second argument must be string")
			}
			new, ok := args[2].(string)
			if !ok {
				return nil, fmt.Errorf("replace: third argument must be string")
			}
			return strings.ReplaceAll(text, old, new), nil
		},
		"trim": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("trim: expected 1 argument, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("trim: argument must be string")
			}
			return strings.TrimSpace(text), nil
		},
		"length": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("length: expected 1 argument, got %d", len(args))
			}
			switch v := args[0].(type) {
			case string:
				return len(v), nil
			case []interface{}:
				return len(v), nil
			default:
				return nil, fmt.Errorf("length: argument must be string or array")
			}
		},
		"toUpper": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("toUpper: expected 1 argument, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("toUpper: argument must be string")
			}
			return strings.ToUpper(text), nil
		},
		"toLower": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("toLower: expected 1 argument, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("toLower: argument must be string")
			}
			return strings.ToLower(text), nil
		},
		"append": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("append: expected 2 arguments, got %d", len(args))
			}
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("append: first argument must be array")
			}
			return append(arr, args[1]), nil
		},
		"remove": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("remove: expected 2 arguments, got %d", len(args))
			}
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("remove: first argument must be array")
			}
			idx, ok := args[1].(int)
			if !ok {
				return nil, fmt.Errorf("remove: second argument must be integer")
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("remove: index out of range")
			}
			return append(arr[:idx], arr[idx+1:]...), nil
		},
		"random": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("random: expected 2 arguments, got %d", len(args))
			}
			min, ok := args[0].(int)
			if !ok {
				return nil, fmt.Errorf("random: first argument must be integer")
			}
			max, ok := args[1].(int)
			if !ok {
				return nil, fmt.Errorf("random: second argument must be integer")
			}
			if min > max {
				return nil, fmt.Errorf("random: min must be <= max")
			}
			now := time.Now().UnixNano()
			seed := now % 1000000
			result := min + int(seed%(int64(max-min+1)))
			return result, nil
		},
		"max": func(args []interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("max: expected at least 2 arguments")
			}
			maxVal := args[0]
			for _, arg := range args[1:] {
				cmp, err := compareValues(arg, maxVal)
				if err != nil {
					return nil, err
				}
				if cmp > 0 {
					maxVal = arg
				}
			}
			return maxVal, nil
		},
		"min": func(args []interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("min: expected at least 2 arguments")
			}
			minVal := args[0]
			for _, arg := range args[1:] {
				cmp, err := compareValues(arg, minVal)
				if err != nil {
					return nil, err
				}
				if cmp < 0 {
					minVal = arg
				}
			}
			return minVal, nil
		},
		"sqrt": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sqrt: expected 1 argument, got %d", len(args))
			}
			var val float64
			switch v := args[0].(type) {
			case int:
				val = float64(v)
			case float64:
				val = v
			default:
				return nil, fmt.Errorf("sqrt: argument must be number")
			}
			if val < 0 {
				return nil, fmt.Errorf("sqrt: cannot take square root of negative number")
			}
			result := val / 2
			for i := 0; i < 100; i++ {
				result = (result + val/result) / 2
			}
			return result, nil
		},
		"pow": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("pow: expected 2 arguments, got %d", len(args))
			}
			var base float64
			switch v := args[0].(type) {
			case int:
				base = float64(v)
			case float64:
				base = v
			default:
				return nil, fmt.Errorf("pow: first argument must be number")
			}
			var exp float64
			switch v := args[1].(type) {
			case int:
				exp = float64(v)
			case float64:
				exp = v
			default:
				return nil, fmt.Errorf("pow: second argument must be number")
			}
			result := 1.0
			for i := 0; i < int(exp); i++ {
				result *= base
			}
			if exp == float64(int(exp)) {
				return result, nil
			}
			return nil, fmt.Errorf("pow: only integer exponents supported")
		},
		"getTime": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getTime: expected 0 arguments")
			}
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
		"getYear": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getYear: expected 0 arguments")
			}
			return time.Now().Year(), nil
		},
		"getMonth": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getMonth: expected 0 arguments")
			}
			return int(time.Now().Month()), nil
		},
		"getOS": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getOS: expected 0 arguments")
			}
			return os.Getenv("OS"), nil
		},
		"getHostname": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getHostname: expected 0 arguments")
			}
			hostname, err := os.Hostname()
			if err != nil {
				return "", err
			}
			return hostname, nil
		},
		"getArgs": func(args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("getArgs: expected 0 arguments")
			}
			argsSlice := os.Args[1:]
			result := make([]interface{}, len(argsSlice))
			for i, a := range argsSlice {
				result[i] = a
			}
			return result, nil
		},
		"hasFlag": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hasFlag: expected 1 argument, got %d", len(args))
			}
			flag, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("hasFlag: argument must be string")
			}
			for _, arg := range os.Args[1:] {
				if arg == flag {
					return true, nil
				}
			}
			return false, nil
		},
		"httpGet": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("httpGet: expected 1 argument, got %d", len(args))
			}
			url, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("httpGet: argument must be string")
			}
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return string(body), nil
		},
		"httpPost": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("httpPost: expected 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("httpPost: first argument must be string")
			}
			data, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("httpPost: second argument must be string")
			}
			resp, err := http.Post(url, "application/json", strings.NewReader(data))
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return string(body), nil
		},
		"jsonParse": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("jsonParse: expected 1 argument, got %d", len(args))
			}
			jsonStr, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("jsonParse: argument must be string")
			}
			var result interface{}
			err := json.Unmarshal([]byte(jsonStr), &result)
			if err != nil {
				return nil, err
			}
			if m, ok := result.(map[string]interface{}); ok {
				return m, nil
			}
			if arr, ok := result.([]interface{}); ok {
				return arr, nil
			}
			return result, nil
		},
		"md5": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("md5: expected 1 argument, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("md5: argument must be string")
			}
			hash := md5.Sum([]byte(text))
			return hex.EncodeToString(hash[:]), nil
		},
		"sha256": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sha256: expected 1 argument, got %d", len(args))
			}
			text, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("sha256: argument must be string")
			}
			hash := sha256.Sum256([]byte(text))
			return hex.EncodeToString(hash[:]), nil
		},
		"regexFind": func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("regexFind: expected 2 arguments, got %d", len(args))
			}
			pattern, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("regexFind: first argument must be string")
			}
			text, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("regexFind: second argument must be string")
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			matches := re.FindAllString(text, -1)
			result := make([]interface{}, len(matches))
			for i, m := range matches {
				result[i] = m
			}
			return result, nil
		},
		"regexReplace": func(args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("regexReplace: expected 3 arguments, got %d", len(args))
			}
			pattern, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("regexReplace: first argument must be string")
			}
			text, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("regexReplace: second argument must be string")
			}
			replacement, ok := args[2].(string)
			if !ok {
				return nil, fmt.Errorf("regexReplace: third argument must be string")
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			return re.ReplaceAllString(text, replacement), nil
		},
		"sleep": func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sleep: expected 1 argument, got %d", len(args))
			}
			ms, ok := args[0].(int)
			if !ok {
				return nil, fmt.Errorf("sleep: argument must be integer")
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return nil, nil
		},
		"sendEmail": func(args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("sendEmail: expected 3 arguments, got %d", len(args))
			}
			to, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("sendEmail: first argument must be string")
			}
			subject, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("sendEmail: second argument must be string")
			}
			body, ok := args[2].(string)
			if !ok {
				return nil, fmt.Errorf("sendEmail: third argument must be string")
			}
			fmt.Fprintf(output, "Email sent to %s: %s - %s\n", to, subject, body)
			return nil, nil
		},
	}

	fn, ok := builtins[name]
	return fn, ok
}

func compareValues(a, b interface{}) (int, error) {
	switch va := a.(type) {
	case int:
		vb, ok := b.(int)
		if !ok {
			return 0, fmt.Errorf("cannot compare int with %T", b)
		}
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	case float64:
		vb, ok := b.(float64)
		if !ok {
			return 0, fmt.Errorf("cannot compare float with %T", b)
		}
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	case string:
		vb, ok := b.(string)
		if !ok {
			return 0, fmt.Errorf("cannot compare string with %T", b)
		}
		if va < vb {
			return -1, nil
		} else if va > vb {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot compare type %T", a)
	}
}

func (i *Interpreter) handleInclude(n *IncludeStatement) (interface{}, error) {
	filename := n.Filename

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
		return nil, fmt.Errorf("library not found: %s (searched in: %v)", filename, searchPaths)
	}

	oldPath := currentFilePath
	if currentFilePath == "" {
		currentFilePath = filename
	}

	lexer := NewLexer(string(fileContent))
	tokens, err := lexer.Tokenize()
	if err != nil {
		currentFilePath = oldPath
		return nil, fmt.Errorf("error tokenizing library %s: %v", filename, err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		currentFilePath = oldPath
		return nil, fmt.Errorf("error parsing library %s: %v", filename, err)
	}

	_, err = i.Evaluate(ast)
	if err != nil {
		currentFilePath = oldPath
		return nil, fmt.Errorf("error executing library %s: %v", filename, err)
	}

	currentFilePath = oldPath
	return nil, nil
}

func (i *Interpreter) Evaluate(node Node) (interface{}, error) {
	switch n := node.(type) {
	case *IncludeStatement:
		return i.handleInclude(n)
	case *CommentStatement:
		return nil, nil
	case *NumberNode:
		return n.Value, nil
	case *FloatNode:
		return n.Value, nil
	case *StringNode:
		return n.Value, nil
	case *BooleanNode:
		return n.Value, nil
	case *VariableNode:
		if val, ok := i.getVar(n.Name); ok {
			return val, nil
		}
		return nil, fmt.Errorf("variable not defined: %s", n.Name)
	case *ArrayNode:
		elements := make([]interface{}, len(n.Elements))
		for idx, elem := range n.Elements {
			val, err := i.Evaluate(elem)
			if err != nil {
				return nil, err
			}
			elements[idx] = val
		}
		return elements, nil
	case *ArrayIndexNode:
		arr, ok := i.getVar(n.Name)
		if !ok {
			return nil, fmt.Errorf("array not defined: %s", n.Name)
		}
		arrSlice, ok := arr.([]interface{})
		if !ok {
			return nil, fmt.Errorf("variable %s is not an array", n.Name)
		}
		idxVal, err := i.Evaluate(n.Index)
		if err != nil {
			return nil, err
		}
		idx, ok := idxVal.(int)
		if !ok {
			return nil, fmt.Errorf("array index must be integer")
		}
		if idx < 0 || idx >= len(arrSlice) {
			return nil, fmt.Errorf("array index out of bounds: %d", idx)
		}
		return arrSlice[idx], nil
	case *BinaryOpNode:
		left, err := i.Evaluate(n.Left)
		if err != nil {
			return nil, err
		}
		right, err := i.Evaluate(n.Right)
		if err != nil {
			return nil, err
		}
		return i.evaluateBinaryOp(left, n.Op, right)
	case *TypedDeclaration:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return nil, err
		}
		if err := i.checkType(n.Type, value); err != nil {
			return nil, err
		}
		i.setVar(n.Name, value)
		i.setType(n.Name, n.Type)
		return nil, nil
	case *AssignmentStatement:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return nil, err
		}
		if typ, ok := i.getType(n.Name); ok {
			if err := i.checkType(typ, value); err != nil {
				return nil, err
			}
			i.setVar(n.Name, value)
		} else {
			return nil, fmt.Errorf("variable not declared: %s", n.Name)
		}
		return nil, nil
	case *ArrayAssignmentStatement:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return nil, err
		}
		arr, ok := i.getVar(n.Name)
		if !ok {
			return nil, fmt.Errorf("array not defined: %s", n.Name)
		}
		arrSlice, ok := arr.([]interface{})
		if !ok {
			return nil, fmt.Errorf("variable %s is not an array", n.Name)
		}
		idxVal, err := i.Evaluate(n.Index)
		if err != nil {
			return nil, err
		}
		idx, ok := idxVal.(int)
		if !ok {
			return nil, fmt.Errorf("array index must be integer")
		}
		if idx < 0 || idx >= len(arrSlice) {
			return nil, fmt.Errorf("array index out of bounds: %d", idx)
		}
		arrSlice[idx] = value
		i.setVar(n.Name, arrSlice)
		return nil, nil
	case *PrintStatement:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return nil, err
		}
		fmt.Fprintln(output, value)
		return nil, nil
	case *IfStatement:
		cond, err := i.Evaluate(n.Condition)
		if err != nil {
			return nil, err
		}
		if i.isTruthy(cond) {
			for _, stmt := range n.ThenBlock {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return nil, err
				}
			}
		} else {
			for _, stmt := range n.ElseBlock {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	case *WhileStatement:
		condVal, err := i.Evaluate(n.Condition)
		if err != nil {
			return nil, err
		}
		for i.isTruthy(condVal) {
			for _, stmt := range n.Body {
				_, err := i.Evaluate(stmt)
				if err != nil {
					return nil, err
				}
			}
			condVal, err = i.Evaluate(n.Condition)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *HelpStatement:
		countryVal, err := i.Evaluate(n.Country)
		if err != nil {
			return nil, err
		}
		country, ok := countryVal.(string)
		if !ok {
			return nil, fmt.Errorf("help argument must be a string")
		}
		i.showHelp(country)
		return nil, nil
	case *OrientationStatement:
		if isFileExecution {
			i.runOrientationDemo()
		} else {
			i.runOrientationTest()
		}
		return nil, nil
	case *FunctionDeclaration:
		i.functions[n.Name] = n
		if n.Exported {
			i.exportedFuncs[n.Name] = n
		}
		return nil, nil
	case *FunctionCall:
		if fn, ok := i.getBuiltinFunction(n.Name); ok {
			args := make([]interface{}, len(n.Args))
			for idx, arg := range n.Args {
				val, err := i.Evaluate(arg)
				if err != nil {
					return nil, err
				}
				args[idx] = val
			}
			return fn(args)
		}

		fn, ok := i.functions[n.Name]
		if !ok {
			return nil, fmt.Errorf("function not defined: %s", n.Name)
		}

		if len(fn.Params) != len(n.Args) {
			return nil, fmt.Errorf("argument count mismatch in call to %s: expected %d, got %d",
				n.Name, len(fn.Params), len(n.Args))
		}

		argValues := make([]interface{}, len(n.Args))
		for idx, arg := range n.Args {
			val, err := i.Evaluate(arg)
			if err != nil {
				return nil, err
			}
			argValues[idx] = val
		}

		i.pushFrame()

		for idx, param := range fn.Params {
			i.setVar(param, argValues[idx])
		}

		i.returnFlag = false
		i.returnValue = nil

		var lastResult interface{}
		var err error
		for _, stmt := range fn.Body {
			lastResult, err = i.Evaluate(stmt)
			if err != nil {
				i.popFrame()
				return nil, err
			}
			if i.returnFlag {
				break
			}
		}

		result := i.returnValue
		if result == nil {
			result = lastResult
		}

		i.popFrame()
		return result, nil
	case *ReturnStatement:
		value, err := i.Evaluate(n.Value)
		if err != nil {
			return nil, err
		}
		i.returnValue = value
		i.returnFlag = true
		return value, nil
	case *TryCatchStatement:
		var err error
		for _, stmt := range n.TryBlock {
			_, err = i.Evaluate(stmt)
			if err != nil {
				break
			}
		}

		if err != nil {
			i.setVar("error", err.Error())
			for _, stmt := range n.CatchBlock {
				_, catchErr := i.Evaluate(stmt)
				if catchErr != nil {
					return nil, catchErr
				}
			}
			return nil, nil
		}
		return nil, nil
	case *ExpressionStatement:
		val, err := i.Evaluate(n.Expr)
		if err != nil {
			return nil, err
		}
		return val, nil
	case *Program:
		for _, stmt := range n.Statements {
			_, err := i.Evaluate(stmt)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	return nil, nil
}

func (i *Interpreter) checkType(typ string, val interface{}) error {
	switch typ {
	case "lesbian":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("type mismatch: expected string for lesbian, got %T", val)
		}
	case "gay":
		if _, ok := val.(int); !ok {
			return fmt.Errorf("type mismatch: expected int for gay, got %T", val)
		}
	case "trans":
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("type mismatch: expected float for trans, got %T", val)
		}
	case "nonbinary":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("type mismatch: expected bool for nonbinary, got %T", val)
		}
	default:
		return fmt.Errorf("unknown type: %s", typ)
	}
	return nil
}

func (i *Interpreter) toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		var parts []string
		for _, elem := range v {
			parts = append(parts, i.toString(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (i *Interpreter) evaluateBinaryOp(left interface{}, op string, right interface{}) (interface{}, error) {
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
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		cmp, err := compareValues(left, right)
		if err != nil {
			return nil, err
		}
		return cmp < 0, nil
	case ">":
		cmp, err := compareValues(left, right)
		if err != nil {
			return nil, err
		}
		return cmp > 0, nil
	case "<=":
		cmp, err := compareValues(left, right)
		if err != nil {
			return nil, err
		}
		return cmp <= 0, nil
	case ">=":
		cmp, err := compareValues(left, right)
		if err != nil {
			return nil, err
		}
		return cmp >= 0, nil
	case "&&":
		return i.isTruthy(left) && i.isTruthy(right), nil
	case "||":
		return i.isTruthy(left) || i.isTruthy(right), nil
	default:
		return nil, fmt.Errorf("unknown operator: %s", op)
	}
}

func (i *Interpreter) add(a, b interface{}) (interface{}, error) {
	if (isInt(a) || isFloat(a)) && (isInt(b) || isFloat(b)) {
		va, vb, err := i.toFloat(a, b)
		if err != nil {
			return nil, err
		}
		if isInt(a) && isInt(b) {
			return int(va) + int(vb), nil
		}
		return va + vb, nil
	}
	return i.toString(a) + i.toString(b), nil
}

func isInt(v interface{}) bool {
	_, ok := v.(int)
	return ok
}
func isFloat(v interface{}) bool {
	_, ok := v.(float64)
	return ok
}

func (i *Interpreter) subtract(a, b interface{}) (interface{}, error) {
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return nil, err
	}
	if isInt(a) && isInt(b) {
		return int(va) - int(vb), nil
	}
	return va - vb, nil
}

func (i *Interpreter) multiply(a, b interface{}) (interface{}, error) {
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return nil, err
	}
	if isInt(a) && isInt(b) {
		return int(va) * int(vb), nil
	}
	return va * vb, nil
}

func (i *Interpreter) divide(a, b interface{}) (interface{}, error) {
	va, vb, err := i.toFloat(a, b)
	if err != nil {
		return nil, err
	}
	if vb == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	if isInt(a) && isInt(b) {
		return int(va) / int(vb), nil
	}
	return va / vb, nil
}

func (i *Interpreter) modulo(a, b interface{}) (interface{}, error) {
	va, ok := a.(int)
	if !ok {
		return nil, fmt.Errorf("modulo requires integer operands")
	}
	vb, ok := b.(int)
	if !ok {
		return nil, fmt.Errorf("modulo requires integer operands")
	}
	if vb == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return va % vb, nil
}

func (i *Interpreter) toFloat(a, b interface{}) (float64, float64, error) {
	var va, vb float64
	switch v := a.(type) {
	case int:
		va = float64(v)
	case float64:
		va = v
	default:
		return 0, 0, fmt.Errorf("cannot convert to number: %T", a)
	}
	switch v := b.(type) {
	case int:
		vb = float64(v)
	case float64:
		vb = v
	default:
		return 0, 0, fmt.Errorf("cannot convert to number: %T", b)
	}
	return va, vb, nil
}

func (i *Interpreter) isTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	default:
		return val != nil
	}
}

// ---------- Реализация команд ----------

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
	case *NumberNode, *FloatNode, *StringNode, *BooleanNode, *VariableNode, *BinaryOpNode, *FunctionCall, *ArrayNode, *ArrayIndexNode:
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
		fmt.Fprintln(output, val)
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

// ---------- Функция runExample ----------
func runExample() {
	program := `
		@ Пример продвинутой программы на LGBTScript Language
		
		@ Импорт библиотек
		#intersex "math.rainbow"
		#intersex "io.rainbow"
		
		@ Экспортируемая функция
		export rainbow processData(data) {
			comingout "Processing: " + data;
			return "Processed: " + data;
		}
		
		@ Основная программа
		rainbow main() {
			@ Работа с файлами
			lesbian config = readFile("config.json");
			comingout "Config loaded: " + config;
			
			@ Парсинг JSON
			lesbian settings = jsonParse(config);
			comingout "API URL: " + settings["api_url"];
			
			@ HTTP запрос
			try {
				lesbian response = httpGet(settings["api_url"]);
				comingout "Response received";
				writeFile("response.txt", response);
			} catch {
				comingout "Error: " + error;
			}
			
			@ Работа с массивами
			gay numbers = [1, 2, 3, 4, 5];
			append(numbers, 6);
			comingout "Numbers: " + numbers;
			comingout "First element: " + numbers[0];
			
			@ Строковые операции
			lesbian text = "  Hello World  ";
			lesbian clean = trim(text);
			lesbian upper = toUpper(clean);
			comingout "Upper: " + upper;
			
			@ Регулярные выражения
			gay matches = regexFind("\\d+", "Phone: 123-456-7890");
			comingout "Found numbers: " + matches;
			
			@ Криптография
			lesbian hash = md5("password");
			comingout "MD5: " + hash;
			
			@ Системная информация
			lesbian os = getOS();
			lesbian host = getHostname();
			comingout "OS: " + os + ", Host: " + host;
			
			@ Аргументы командной строки
			gay args = getArgs();
			comingout "Arguments: " + args;
			
			nonbinary verbose = hasFlag("--verbose");
			if (verbose) {
				comingout "Verbose mode enabled";
			}
			
			return 0;
		}
		
		@ Запуск
		main();
	`

	fmt.Fprintln(output, "=== LGBTScript - Расширенный пример ===")
	fmt.Fprintln(output, "Исходная программа:")
	fmt.Fprintln(output, program)
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
	_, err = interpreter.Evaluate(ast)
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
	example := flag.Bool("example", false, "показать расширенный пример")
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
		_, err = interpreter.Evaluate(ast)
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
		_, err = interpreter.Evaluate(ast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
			os.Exit(1)
		}
		return
	}

	runDemoOnly()
}

func runDemoOnly() {
	program := `
		@ Демонстрация всех новых возможностей LGBTScript
		
		lesbian x;
		x = "1";
		comingout x;
		
		gay d;
		comingout d;
		comingout 4;
		
		@ Работа с файлами
		writeFile("test.txt", "Hello Rainbow!");
		lesbian content = readFile("test.txt");
		comingout "File content: " + content;
		
		@ Массивы
		gay numbers = [1, 2, 3, 4, 5];
		append(numbers, 6);
		comingout "Array: " + numbers;
		comingout "Length: " + length(numbers);
		
		@ Строки
		lesbian text = "  Hello World  ";
		comingout "Trimmed: " + trim(text);
		comingout "Upper: " + toUpper(text);
		comingout "Length: " + length(text);
		
		@ Регулярные выражения
		lesbian phone = "Phone: 123-456-7890";
		gay matches = regexFind("\\d+", phone);
		comingout "Phone numbers: " + matches;
		
		@ HTTP
		lesbian url = "https://api.github.com";
		try {
			lesbian response = httpGet(url);
			comingout "HTTP response length: " + length(response);
		} catch {
			comingout "HTTP error: " + error;
		}
		
		@ Криптография
		lesbian hash = md5("password");
		comingout "MD5 hash: " + hash;
		
		@ Система
		lesbian os = getOS();
		lesbian host = getHostname();
		comingout "OS: " + os;
		comingout "Host: " + host;
		
		@ Математика
		gay randomNum = random(1, 100);
		comingout "Random: " + randomNum;
		comingout "Max of 10, 20, 30: " + max(10, 20, 30);
		comingout "Min of 10, 20, 30: " + min(10, 20, 30);
		
		@ Функции с экспортом
		export rainbow greet(name) {
			return "Hello, " + name + "!";
		}
		
		comingout greet("Alice");
		
		@ Аргументы командной строки
		gay args = getArgs();
		comingout "Arguments: " + args;
	`

	fmt.Fprintln(output, "=== LGBTScript - Демонстрация всех возможностей ===")
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
	_, err = interpreter.Evaluate(ast)
	if err != nil {
		fmt.Fprintln(output, "Ошибка выполнения:", err)
	}
}