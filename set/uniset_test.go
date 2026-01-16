package set

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var setAddTestCases = []struct {
	r      []rune
	expect string
}{
	{[]rune{}, "{}"},
	{[]rune{0}, "{0x0000..0x0000}"},
	{[]rune{1, 0}, "{0x0000..0x0001}"},
	{[]rune{1, 0, 1, 1, 0, 1}, "{0x0000..0x0001}"},
	{[]rune{0, 2, 1}, "{0x0000..0x0002}"},
	{[]rune{5, 0, 3, 1, 4}, "{0x0000..0x0001,0x0003..0x0005}"},
	{[]rune{8, 5, 1}, "{0x0001..0x0001,0x0005..0x0005,0x0008..0x0008}"},
	{[]rune{8, 5, 3}, "{0x0003..0x0003,0x0005..0x0005,0x0008..0x0008}"},
	{[]rune{8, 5, 4}, "{0x0004..0x0005,0x0008..0x0008}"},
	{[]rune{8, 5, 5}, "{0x0005..0x0005,0x0008..0x0008}"},
	{[]rune{8, 5, 6}, "{0x0005..0x0006,0x0008..0x0008}"},
	{[]rune{8, 5, 7}, "{0x0005..0x0005,0x0007..0x0008}"},
	{[]rune{8, 5, 8}, "{0x0005..0x0005,0x0008..0x0008}"},
	{[]rune{8, 5, 9}, "{0x0005..0x0005,0x0008..0x0009}"},
	{[]rune{8, 5, 10}, "{0x0005..0x0005,0x0008..0x0008,0x000a..0x000a}"},
	{[]rune{8, 7, 6, 1, 0, 3, 2, 15, 10},
		"{0x0000..0x0003,0x0006..0x0008,0x000a..0x000a,0x000f..0x000f}"},
}

func stringify(runes []rune) string {
	sb := strings.Builder{}
	sb.WriteRune('[')
	for i, r := range runes {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(fmt.Sprintf("%d", r))
	}
	sb.WriteRune(']')
	return sb.String()
}

func TestBase(t *testing.T) {
	set := NewUniSet()
	assert.Equal(t, 0, len(slices.Collect(set.Range)))
	assert.True(t, set.Add('a'))
	assert.True(t, set.Add('b'))
	assert.False(t, set.Add('a')) // already added
	assert.True(t, set.Add('f'))

	assert.True(t, set.Find('a'))
	assert.True(t, set.Find('b'))
	assert.False(t, set.Find('c'))
	assert.True(t, set.Find('f'))

	set = NewUniSet('a', 'b', 'c', 'e', 'f')
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x}", 'a', 'c', 'e', 'f'), set.String())
	runes := slices.Collect(set.Range)
	assert.Equal(t, 2, len(runes))
	assert.Equal(t, 'a', runes[0].First)
	assert.Equal(t, 'c', runes[0].Last)
	assert.Equal(t, 'e', runes[1].First)
	assert.Equal(t, 'f', runes[1].Last)

	assert.True(t, set.Add('d'))
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'f'), set.String())
	runes = slices.Collect(set.Range)
	assert.Equal(t, 1, len(runes))
	assert.Equal(t, 'a', runes[0].First)
	assert.Equal(t, 'f', runes[0].Last)

	set = UniSet{}
	set.AddRange(RuneRange{'a', 'e'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'e'), set.String())
	assert.True(t, set.Find('a'))
	assert.True(t, set.Find('b'))
	assert.True(t, set.Find('c'))
	assert.True(t, set.Find('d'))
	assert.True(t, set.Find('e'))
	assert.False(t, set.Find('f'))
	set.RemoveRange(RuneRange{'c', 'e'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'b'), set.String())

	set = UniSet{}
	set.AddRange(RuneRange{'c', 'a'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'c'), set.String())
	assert.True(t, set.Remove('a'))
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'b', 'c'), set.String())

	set = NewUniSet('a', 'b', 'c', 'e', 'f')
	other := NewUniSet('a', 'c', 'e', 'g')
	set.RemoveSet(&other)
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x}", 'b', 'b', 'f', 'f'), set.String())
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x,0x%04x..0x%04x,0x%04x..0x%04x}",
		'a', 'a', 'c', 'c', 'e', 'e', 'g', 'g'), other.String())

	set = NewUniSet()
	assert.False(t, set.Add(-1)) // ignore invalid rune

	// intersection
	set = NewUniSet('a', 'b', 'c', 'e', 'f')
	other = NewUniSet('a', 'c', 'e', 'g')
	set = set.AndSet(&other)
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x,0x%04x..0x%04x}",
		'a', 'a', 'c', 'c', 'e', 'e'), set.String())

	set = NewUniSet('a', 'b', 'c', 'd', 'f')
	other = NewUniSet('g', 'h', 'i', 'j', 'k')
	set = set.AndSet(&other)
	assert.Equal(t, "{}", set.String())
}

func TestAdd(t *testing.T) {
	for i, testCase := range setAddTestCases {
		set := UniSet{}
		for _, v := range testCase.r {
			set.Add(v)
		}
		assert.Equal(t, testCase.expect, set.String(),
			fmt.Sprintf("index=%d, %s", i, stringify(testCase.r)))
	}
}

func TestSample(t *testing.T) {
	var table = []struct {
		first int
		last  int
	}{
		{0xE000, 0xF8FF},
		{0xF0000, 0xFFFFD},
		{0x100000, 0x10FFFD},
	}
	builder := UniSetBuilder{}
	for _, v := range table {
		builder.AddRange(RuneRange{rune(v.first), rune(v.last)})
	}
	set := builder.Build()
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	sampled := set.Sample(rnd, set.Len()+10)
	assert.Equal(t, set.Len(), sampled.Len(), "sample size")
	for r := range sampled.Iter {
		assert.True(t, set.Find(r), fmt.Sprintf("rune U+%04x", r))
	}

	sampled = set.Sample(rnd, -122)
	assert.Equal(t, 0, sampled.Len(), "negative sample size")
}
