package op

import (
	"fmt"
	"strings"
)

type String struct {
	str   string
	runes []rune
}

func NewString(runes []rune) String {
	return String{str: string(runes), runes: runes}
}

func (s String) String() string {
	return s.str
}

func (s String) Runes() []rune {
	return s.runes
}

func (s String) Utf8Escaped() string {
	v := s.String()
	sb := strings.Builder{}
	for i := 0; i < len(v); i++ {
		b := v[i]
		sb.WriteString(fmt.Sprintf("\\x%02X", b))
	}
	return sb.String()
}
