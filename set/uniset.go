package set

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// RuneInterval code point interval
type RuneInterval struct {
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
	runes []rune //TODO: use rune interval for large data
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

func (u *UniSetBuilder) AddInterval(interval RuneInterval) {
	first := min(interval.First, interval.Last)
	first = max(0, first)
	last := max(interval.First, interval.Last)
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
	builder.AddInterval(RuneInterval{0, utf8.MaxRune})
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

func (u *UniSet) AddInterval(interval RuneInterval) {
	builder := TakeFromSet(u)
	builder.AddInterval(interval)
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

func (u *UniSet) RemoveInterval(interval RuneInterval) {
	first := min(interval.First, interval.Last)
	last := max(interval.First, interval.Last)
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

func (u *UniSet) Interval(yield func(interval RuneInterval) bool) {
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
		if !yield(RuneInterval{first, last}) {
			return
		}
	}
}

func (u *UniSet) String() string {
	sb := strings.Builder{}
	sb.WriteRune('{')
	sb.Grow(len(u.runes) / 4)
	c := 0
	for interval := range u.Interval {
		if c > 0 {
			sb.WriteRune(',')
		}
		c += 1
		sb.WriteString(fmt.Sprintf("0x%04x..0x%04x", interval.First, interval.Last))
	}
	sb.WriteRune('}')
	return sb.String()
}
