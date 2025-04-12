package op

import (
	"errors"
	"fmt"
	"regexp"
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

func (p *Parser) error(msg string) {
	p.err = errors.New(msg)
	panic(p.err)
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
	node = p.parseUnionOrDiff()
	if p.err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseUnionOrDiff() Node {
	return nil
}
