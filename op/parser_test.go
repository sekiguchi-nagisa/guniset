package op

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var lexerTestCases = []struct {
	src    string
	tokens []Token
}{
	{"1234", []Token{{TokenRune, "1234"}}},
	{" 1234s", []Token{{TokenSpace, " "}, {TokenRune, "1234"}, {TokenId, "s"}}},
	{"1234+  cat:eee, five ", []Token{
		{TokenRune, "1234"}, {TokenPlus, "+"},
		{TokenSpace, "  "}, {TokenId, "cat"},
		{TokenColon, ":"}, {TokenId, "eee"},
		{TokenComma, ","}, {TokenSpace, " "},
		{TokenId, "five"}, {TokenSpace, " "},
	}},
	{"0..U+f", []Token{{TokenRune, "0"}, {TokenRange, ".."}, {TokenRune, "U+f"}}},
	{"0..f", []Token{{TokenRune, "0"}, {TokenRange, ".."}, {TokenId, "f"}}},
	{"0..0f", []Token{{TokenRune, "0"}, {TokenRange, ".."}, {TokenRune, "0f"}}},
	{"-124", []Token{{TokenMinus, "-"}, {TokenRune, "124"}}},
	{"U+(455)", []Token{{TokenId, "U"}, {TokenPlus, "+"},
		{TokenLParen, "("}, {TokenRune, "455"}, {TokenRParen, ")"}}},
}

func TestLexer(t *testing.T) {
	for _, testCase := range lexerTestCases {
		actual, err := Tokenize([]byte(testCase.src))
		assert.Nil(t, err)
		assert.Equal(t, testCase.tokens, actual, testCase.src)
	}
}
