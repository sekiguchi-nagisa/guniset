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

func TestParserPrimary(t *testing.T) {
	node, err := NewParser().Run([]byte(" 1234 "))
	assert.Nil(t, err)
	assert.IsType(t, &IntervalNode{}, node)
	assert.Equal(t, rune(0x1234), node.(*IntervalNode).interval.First)
	assert.Equal(t, rune(0x1234), node.(*IntervalNode).interval.Last)

	node, err = NewParser().Run([]byte(" 1234 .. U+FFF0 "))
	assert.Nil(t, err)
	assert.IsType(t, &IntervalNode{}, node)
	assert.Equal(t, rune(0x1234), node.(*IntervalNode).interval.First)
	assert.Equal(t, rune(0xFFF0), node.(*IntervalNode).interval.Last)

	node, err = NewParser().Run([]byte("cat : Lu"))
	assert.Nil(t, err)
	assert.IsType(t, &GeneralCategoryNode{}, node)
	assert.Equal(t, 1, len(node.(*GeneralCategoryNode).properties))
	assert.Equal(t, CAT_Lu, node.(*GeneralCategoryNode).properties[0])

	node, err = NewParser().Run([]byte("cat : Lu,Cn,  Lu ,   Mn , Cn   "))
	assert.Nil(t, err)
	assert.IsType(t, &GeneralCategoryNode{}, node)
	assert.Equal(t, 3, len(node.(*GeneralCategoryNode).properties))
	assert.Equal(t, CAT_Lu, node.(*GeneralCategoryNode).properties[0])
	assert.Equal(t, CAT_Mn, node.(*GeneralCategoryNode).properties[1])
	assert.Equal(t, CAT_Cn, node.(*GeneralCategoryNode).properties[2])

	node, err = NewParser().Run([]byte("eaw: W"))
	assert.Nil(t, err)
	assert.IsType(t, &EastAsianWidthNode{}, node)
	assert.Equal(t, 1, len(node.(*EastAsianWidthNode).properties))
	assert.Equal(t, EAW_W, node.(*EastAsianWidthNode).properties[0])

	node, err = NewParser().Run([]byte("eaw \n: N  ,  W\n,  N  ,F \n \n \t"))
	assert.Nil(t, err)
	assert.IsType(t, &EastAsianWidthNode{}, node)
	assert.Equal(t, 3, len(node.(*EastAsianWidthNode).properties))
	assert.Equal(t, EAW_W, node.(*EastAsianWidthNode).properties[0])
	assert.Equal(t, EAW_F, node.(*EastAsianWidthNode).properties[1])
	assert.Equal(t, EAW_N, node.(*EastAsianWidthNode).properties[2])
}

func TestParserBinary(t *testing.T) {
	node, err := NewParser().Run([]byte("\t \t\n  cat:Zs + 0FeFf  "))
	assert.Nil(t, err)
	assert.IsType(t, &UnionNode{}, node)
	assert.IsType(t, &GeneralCategoryNode{}, node.(*UnionNode).left)
	assert.Equal(t, 1, len(node.(*UnionNode).left.(*GeneralCategoryNode).properties))
	assert.Equal(t, CAT_Zs, node.(*UnionNode).left.(*GeneralCategoryNode).properties[0])
	assert.IsType(t, &IntervalNode{}, node.(*UnionNode).right)
	assert.Equal(t, rune(0xfeff), node.(*UnionNode).right.(*IntervalNode).interval.First)
	assert.Equal(t, rune(0xfeff), node.(*UnionNode).right.(*IntervalNode).interval.Last)

	node, err = NewParser().Run([]byte("(eaw:F + cat:Zs) - (U+1234 + 0FeFf)"))
	assert.Nil(t, err)
	assert.IsType(t, &DiffNode{}, node)
	assert.IsType(t, &UnionNode{}, node.(*DiffNode).left)
	assert.IsType(t, &EastAsianWidthNode{}, node.(*DiffNode).left.(*UnionNode).left)
	assert.Equal(t, 1, len(node.(*DiffNode).left.(*UnionNode).left.(*EastAsianWidthNode).properties))
	assert.Equal(t, EAW_F, node.(*DiffNode).left.(*UnionNode).
		left.(*EastAsianWidthNode).properties[0])
	assert.IsType(t, &GeneralCategoryNode{}, node.(*DiffNode).left.(*UnionNode).right)
	assert.Equal(t, 1, len(node.(*DiffNode).left.(*UnionNode).right.(*GeneralCategoryNode).properties))
	assert.Equal(t, CAT_Zs, node.(*DiffNode).left.(*UnionNode).
		right.(*GeneralCategoryNode).properties[0])

	assert.IsType(t, &UnionNode{}, node.(*DiffNode).right)
	assert.IsType(t, &IntervalNode{}, node.(*DiffNode).right.(*UnionNode).left)
	assert.Equal(t, rune(0x1234), node.(*DiffNode).right.(*UnionNode).left.(*IntervalNode).interval.First)
	assert.Equal(t, rune(0x1234), node.(*DiffNode).right.(*UnionNode).left.(*IntervalNode).interval.Last)
	assert.IsType(t, &IntervalNode{}, node.(*DiffNode).right.(*UnionNode).right)
	assert.Equal(t, rune(0xfeff), node.(*DiffNode).right.(*UnionNode).right.(*IntervalNode).interval.First)
	assert.Equal(t, rune(0xfeff), node.(*DiffNode).right.(*UnionNode).right.(*IntervalNode).interval.Last)
}
