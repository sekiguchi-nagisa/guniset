package set

import (
	"fmt"
	"slices"
	"strings"
)

// RuneInterval code point interval
type RuneInterval struct {
	first rune
	last  rune
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

func (u *UniSet) Add(r rune) bool {
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
	first := min(interval.first, interval.last)
	last := max(interval.first, interval.last)
	for i := first; i <= last; i++ {
		u.Add(i)
	}
}

func (u *UniSet) AddSet(other *UniSet) {
	if other == nil || u == other {
		return
	}
	for _, r := range other.runes {
		u.Add(r)
	}
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
	first := min(interval.first, interval.last)
	last := max(interval.first, interval.last)
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

func (u *UniSet) Find(r rune) bool {
	_, s := slices.BinarySearch(u.runes, r)
	return s
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
		sb.WriteString(fmt.Sprintf("0x%04x..0x%04x", interval.first, interval.last))
	}
	sb.WriteRune('}')
	return sb.String()
}
