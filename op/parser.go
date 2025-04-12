package op

import (
	"errors"
	"fmt"
	"github.com/sekiguchi-nagisa/guniset/set"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type TokenKind -trimprefix Token -linecomment

type TokenKind int

const (
	TokenId     TokenKind = iota // identifier
	TokenRune                    // codePoint
	TokenColon                   // :
	TokenComma                   // ,
	TokenLParen                  // (
	TokenRParen                  // )
	TokenPlus                    // +
	TokenMinus                   // -
	TokenRange                   // ..
	TokenSpace                   // space
)

type Lexeme struct {
	pattern *regexp.Regexp
	kind    TokenKind
}

var lexemes = []Lexeme{
	{regexp.MustCompile(`^U[+][0-9a-fA-F]+`), TokenRune},
	{regexp.MustCompile(`^[0-9][0-9a-fA-F]*`), TokenRune},
	{regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`), TokenId},
	{regexp.MustCompile(`^:`), TokenColon},
	{regexp.MustCompile(`^,`), TokenComma},
	{regexp.MustCompile(`^[(]`), TokenLParen},
	{regexp.MustCompile(`^[)]`), TokenRParen},
	{regexp.MustCompile(`^[+]`), TokenPlus},
	{regexp.MustCompile(`^-`), TokenMinus},
	{regexp.MustCompile(`^[.][.]`), TokenRange},
	{regexp.MustCompile(`^[ \t\n]+`), TokenSpace},
}

type Token struct {
	kind TokenKind
	text string
}

func Tokenize(src []byte) ([]Token, error) {
	tokens := make([]Token, 0)
Next:
	for pos := 0; pos < len(src); {
		buf := src[pos:]
		for _, lexeme := range lexemes {
			r := lexeme.pattern.FindIndex(buf)
			if r == nil {
				continue
			}
			tokens = append(tokens, Token{lexeme.kind, string(src[pos : pos+r[1]])})
			pos += r[1]
			continue Next
		}
		return tokens, fmt.Errorf("invalid token: %s", string(src[pos:]))
	}
	return tokens, nil
}

type Parser struct {
	tokens []Token
	pos    int
	err    error
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) error(msg string) {
	p.err = errors.New(msg)
	panic(p.err)
}

func (p *Parser) hasNext() bool {
	return p.pos < len(p.tokens)
}

func (p *Parser) fetch() *Token {
	if p.hasNext() {
		return &p.tokens[p.pos]
	}
	p.error("unexpected end of token")
	return nil
}

func (p *Parser) consume() {
	p.pos++
	p.skipSpace()
}

func (p *Parser) skipSpace() {
	for p.hasNext() && p.tokens[p.pos].kind == TokenSpace {
		p.pos++
	}
}

func (p *Parser) expect(kind TokenKind) *Token {
	token := p.fetch()
	if token.kind != kind {
		p.error(fmt.Sprintf("token mismatched, expect: %s, actual: %s", kind.String(),
			token.kind.String()))
	}
	p.consume()
	return token
}

func (p *Parser) Run(src []byte) (node Node, err error) {
	tokens, err := Tokenize(src)
	if err != nil {
		return nil, err
	}
	p.tokens = tokens
	p.pos = 0
	p.err = nil
	defer func() {
		recover()
		err = p.err
	}()
	p.skipSpace()
	node = p.parseUnionOrDiff()
	if p.hasNext() {
		p.error(fmt.Sprintf("unexpected token: %s", p.fetch().kind.String()))
	}
	if p.err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parsePropertySeq(consumer func(string)) {
	token := p.expect(TokenId)
	consumer(token.text)
	for p.hasNext() && p.fetch().kind == TokenComma {
		p.consume()
		token = p.expect(TokenId)
		consumer(token.text)
	}
}

func (p *Parser) parseRune() rune {
	s := p.expect(TokenRune).text
	if strings.HasPrefix(s, "U+") {
		s = strings.TrimPrefix(s, "U+")
	}
	v, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		p.error(fmt.Sprintf("invalid rune: %s", err.Error()))
	}
	r := rune(v)
	if !utf8.ValidRune(r) {
		p.error(fmt.Sprintf("out of range rune: %04x", r))
	}
	return r
}

func (p *Parser) parsePrimary() Node {
	switch curKind := p.fetch().kind; curKind {
	case TokenId:
		prefix := p.expect(TokenId)
		if prefix.text == "cat" {
			p.expect(TokenColon)
			var properties []GeneralCategory
			p.parsePropertySeq(func(s string) {
				v, err := ParseGeneralCategory(s)
				if err != nil {
					p.error(err.Error())
				}
				properties = append(properties, v)
			})
			return NewGeneralCategoryNode(properties)
		} else if prefix.text == "eaw" {
			p.expect(TokenColon)
			var properties []EastAsianWidth
			p.parsePropertySeq(func(s string) {
				v, err := ParseEastAsianWidth(s)
				if err != nil {
					p.error(err.Error())
				}
				properties = append(properties, v)
			})
			return NewEastAsianWidthNode(properties)
		} else {
			p.error(fmt.Sprintf("unknown property prefix: %s, must be `cat` or `eaw`", prefix.text))
		}
	case TokenRune:
		first := p.parseRune()
		last := first
		if p.hasNext() && p.fetch().kind == TokenRange {
			p.consume()
			last = p.parseRune()
		}
		return &IntervalNode{interval: set.RuneInterval{First: first, Last: last}}
	case TokenLParen:
		p.consume()
		node := p.parseUnionOrDiff()
		p.expect(TokenRParen)
		return node
	default:
		p.error(fmt.Sprintf("unknown token: %s", curKind.String()))
	}
	return nil
}

func (p *Parser) parseUnionOrDiff() Node {
	left := p.parsePrimary()
	if p.hasNext() {
		switch curKind := p.fetch().kind; curKind {
		case TokenPlus:
			p.consume()
			right := p.parseUnionOrDiff()
			return &UnionNode{left, right}
		case TokenMinus:
			p.consume()
			right := p.parseUnionOrDiff()
			return &DiffNode{left, right}
		default:
		}
	}
	return left
}
