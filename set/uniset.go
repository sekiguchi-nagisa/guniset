package set

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// RuneRange code point range (inclusive, inclusive)
type RuneRange struct {
	First rune
	Last  rune
}

func IsValidRune(r rune) bool {
	return 0 <= r && r <= utf8.MaxRune
}

func IsBmpRune(r rune) bool {
	return 0 <= r && r <= 0xFFFF
}

func IsSupplementaryRune(r rune) bool {
	return r > 0xFFFF
}

func ParseRune(s string) (rune, error) {
	s = strings.TrimPrefix(s, "U+")
	v, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return utf8.RuneError, err
	}
	r := rune(v)
	if !IsValidRune(r) {
		return utf8.RuneError, fmt.Errorf("out of range rune: %04x", r)
	}
	return r, nil
}

// UniSet set structure for Unicode code point
type UniSet struct {
	runes []rune //TODO: use rune range for large data
}

func NewUniSet(runes ...rune) UniSet {
	set := UniSet{}
	for _, r := range runes {
		set.Add(r)
	}
	return set
}

type UniSetBuilder struct {
	runes []rune
}

func TakeFromSet(set *UniSet) UniSetBuilder {
	builder := UniSetBuilder{}
	builder.runes = set.runes
	set.runes = nil
	return builder
}

func (u *UniSetBuilder) Add(r rune) {
	u.runes = append(u.runes, r)
}

func (u *UniSetBuilder) AddRange(runeRange RuneRange) {
	first := min(runeRange.First, runeRange.Last)
	first = max(0, first)
	last := max(runeRange.First, runeRange.Last)
	last = min(last, utf8.MaxRune)
	for i := first; i <= last; i++ {
		u.Add(i)
	}
}

func (u *UniSetBuilder) AddSet(set *UniSet) {
	u.runes = append(u.runes, set.runes...)
}

func (u *UniSetBuilder) BuildRaw() []rune {
	slices.Sort(u.runes)
	u.runes = slices.Compact(u.runes)
	var tmp = u.runes
	u.runes = nil
	return tmp
}

func (u *UniSetBuilder) Build() UniSet {
	return UniSet{runes: u.BuildRaw()}
}

func NewUniSetAll() UniSet {
	builder := UniSetBuilder{}
	builder.AddRange(RuneRange{0, utf8.MaxRune})
	return builder.Build()
}

func (u *UniSet) Add(r rune) bool {
	if !IsValidRune(r) {
		return false
	}
	if u.runes == nil {
		u.runes = []rune{}
	}
	if len(u.runes) == 0 {
		u.runes = append(u.runes, r)
		return true
	}
	i, s := slices.BinarySearch(u.runes, r)
	if s {
		return false
	}
	u.runes = slices.Insert(u.runes, i, r)
	return true
}

func (u *UniSet) AddRange(runeRange RuneRange) {
	builder := TakeFromSet(u)
	builder.AddRange(runeRange)
	u.runes = builder.BuildRaw()
}

func (u *UniSet) AddSet(other *UniSet) {
	if other == nil || u == other {
		return
	}
	builder := TakeFromSet(u)
	builder.AddSet(other)
	u.runes = builder.BuildRaw()
}

func (u *UniSet) Remove(r rune) bool {
	if u.runes == nil || len(u.runes) == 0 {
		return false
	}
	i, s := slices.BinarySearch(u.runes, r)
	if s {
		u.runes = slices.Delete(u.runes, i, i+1)
	}
	return s
}

func (u *UniSet) RemoveRange(runeRange RuneRange) {
	first := min(runeRange.First, runeRange.Last)
	last := max(runeRange.First, runeRange.Last)
	u.runes = slices.DeleteFunc(u.runes, func(r rune) bool {
		return r >= first && r <= last
	})
}

func (u *UniSet) RemoveSet(other *UniSet) {
	if other == nil || u == other {
		return
	}
	u.runes = slices.DeleteFunc(u.runes, func(r rune) bool {
		return other.Find(r)
	})
}

func (u *UniSet) AndSet(other *UniSet) UniSet {
	builder := UniSetBuilder{}
	for _, r := range other.runes {
		if u.Find(r) {
			builder.Add(r)
		}
	}
	return builder.Build()
}

func (u *UniSet) Filter(f func(r rune) bool) {
	u.runes = slices.DeleteFunc(u.runes, func(r rune) bool {
		return !f(r)
	})
}

func (u *UniSet) Find(r rune) bool {
	_, s := slices.BinarySearch(u.runes, r)
	return s
}

func (u *UniSet) Copy() UniSet {
	copied := UniSet{}
	copied.runes = u.runes[0:]
	return copied
}

func (u *UniSet) Range(yield func(runeRange RuneRange) bool) {
	for i := 0; i < len(u.runes); {
		first := u.runes[i]
		last := first
		i += 1
		for i < len(u.runes) {
			cur := u.runes[i]
			if last+1 == cur {
				last = cur
				i += 1
			} else {
				break
			}
		}
		if !yield(RuneRange{first, last}) {
			return
		}
	}
}

func (u UniSet) Iter(yield func(r rune) bool) {
	for _, r := range u.runes {
		if !yield(r) {
			return
		}
	}
}

func (u *UniSet) String() string {
	sb := strings.Builder{}
	sb.WriteRune('{')
	sb.Grow(len(u.runes) / 4)
	c := 0
	for runeRange := range u.Range {
		if c > 0 {
			sb.WriteRune(',')
		}
		c += 1
		sb.WriteString(fmt.Sprintf("0x%04x..0x%04x", runeRange.First, runeRange.Last))
	}
	sb.WriteRune('}')
	return sb.String()
}

func (u *UniSet) Sample(limit int) UniSet {
	runeSet := map[rune]struct{}{}
	limit = min(limit, len(u.runes)/2)
	for len(runeSet) < limit {
		runeSet[u.runes[rand.Intn(len(u.runes))]] = struct{}{}
	}
	builder := UniSetBuilder{}
	for r := range runeSet {
		builder.Add(r)
	}
	return builder.Build()
}
