package set

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"slices"
	"strings"
	"testing"
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
	assert.Equal(t, 0, len(slices.Collect(set.Interval)))
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
	runes := slices.Collect(set.Interval)
	assert.Equal(t, 2, len(runes))
	assert.Equal(t, 'a', runes[0].first)
	assert.Equal(t, 'c', runes[0].last)
	assert.Equal(t, 'e', runes[1].first)
	assert.Equal(t, 'f', runes[1].last)

	assert.True(t, set.Add('d'))
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'f'), set.String())
	runes = slices.Collect(set.Interval)
	assert.Equal(t, 1, len(runes))
	assert.Equal(t, 'a', runes[0].first)
	assert.Equal(t, 'f', runes[0].last)

	set = UniSet{}
	set.AddInterval(RuneInterval{'a', 'e'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'e'), set.String())
	assert.True(t, set.Find('a'))
	assert.True(t, set.Find('b'))
	assert.True(t, set.Find('c'))
	assert.True(t, set.Find('d'))
	assert.True(t, set.Find('e'))
	assert.False(t, set.Find('f'))
	set.RemoveInterval(RuneInterval{'c', 'e'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'b'), set.String())

	set = UniSet{}
	set.AddInterval(RuneInterval{'c', 'a'})
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'a', 'c'), set.String())
	assert.True(t, set.Remove('a'))
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x}", 'b', 'c'), set.String())

	set = NewUniSet('a', 'b', 'c', 'e', 'f')
	other := NewUniSet('a', 'c', 'e', 'g')
	set.RemoveSet(&other)
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x}", 'b', 'b', 'f', 'f'), set.String())
	assert.Equal(t, fmt.Sprintf("{0x%04x..0x%04x,0x%04x..0x%04x,0x%04x..0x%04x,0x%04x..0x%04x}",
		'a', 'a', 'c', 'c', 'e', 'e', 'g', 'g'), other.String())
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
