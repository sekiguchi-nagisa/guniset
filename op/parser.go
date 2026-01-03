package op

import (
	"fmt"
	"regexp"

	"github.com/sekiguchi-nagisa/guniset/set"
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
	TokenNegate                  // !
	TokenPlus                    // +
	TokenMinus                   // -
	TokenMul                     // *
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
	{regexp.MustCompile(`^!`), TokenNegate},
	{regexp.MustCompile(`^[+]`), TokenPlus},
	{regexp.MustCompile(`^-`), TokenMinus},
	{regexp.MustCompile(`^[*]`), TokenMul},
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
	aliasMaps AliasMaps
	tokens    []Token
	pos       int
	err       error
}

func NewParser(maps AliasMaps) *Parser {
	return &Parser{aliasMaps: maps}
}

func syntaxErr(msg string) error {
	return fmt.Errorf("[syntax error] %s", msg)
}

func (p *Parser) error(msg string) {
	p.err = syntaxErr(msg)
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
		return nil, syntaxErr(err.Error())
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
	r, err := set.ParseRune(s)
	if err != nil {
		p.error(err.Error())
	}
	return r
}

func (p *Parser) parsePrimary() Node {
	switch curKind := p.fetch().kind; curKind {
	case TokenId:
		prefix := p.expect(TokenId)
		if IsGeneralCategoryPrefix(prefix.text) {
			p.expect(TokenColon)
			var properties []GeneralCategory
			p.parsePropertySeq(func(s string) {
				v, err := ParseGeneralCategory(s, p.aliasMaps[GeneralCategoryPrefix])
				if err != nil {
					p.error(err.Error())
				}
				properties = append(properties, v)
			})
			return NewGeneralCategoryNode(properties)
		} else if IsEastAsianWidthPrefix(prefix.text) {
			p.expect(TokenColon)
			var properties []EastAsianWidth
			p.parsePropertySeq(func(s string) {
				v, err := ParseEastAsianWidth(s, p.aliasMaps[EastAsianWidthPrefix])
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
		return &RangeNode{runeRange: set.RuneRange{First: first, Last: last}}
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

func (p *Parser) parseComplement() Node {
	if p.fetch().kind == TokenNegate {
		p.consume()
		node := p.parseComplement()
		return &CompNode{node: node}
	}
	return p.parsePrimary()
}

func (p *Parser) parseIntersect() Node {
	left := p.parseComplement()
	for p.hasNext() {
		switch curKind := p.fetch().kind; curKind {
		case TokenMul:
			p.consume()
			right := p.parseComplement()
			left = &IntersectNode{left, right}
			continue
		default:
		}
		break
	}
	return left
}

func (p *Parser) parseUnionOrDiff() Node {
	left := p.parseIntersect()
	for p.hasNext() {
		switch curKind := p.fetch().kind; curKind {
		case TokenPlus:
			p.consume()
			right := p.parseIntersect()
			left = &UnionNode{left, right}
			continue
		case TokenMinus:
			p.consume()
			right := p.parseIntersect()
			left = &DiffNode{left, right}
			continue
		default:
		}
		break
	}
	return left
}
