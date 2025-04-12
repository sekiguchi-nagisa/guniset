package op

import "fmt"

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type GeneralCategory -trimprefix CAT_

type GeneralCategory int

const (
	CAT_Lu GeneralCategory = iota
	CAT_Ll
	CAT_Lt
	CAT_Lm
	CAT_Lo
	CAT_Mn
	CAT_Mc
	CAT_Me
	CAT_Nd
	CAT_Nl
	CAT_No
	CAT_Sm
	CAT_Sc
	CAT_Sk
	CAT_So
	CAT_Pc
	CAT_Pd
	CAT_Ps
	CAT_Pe
	CAT_Pi
	CAT_Pf
	CAT_Po
	CAT_Zs
	CAT_Zl
	CAT_Zp
	CAT_Cc
	CAT_Cf
	CAT_Cs
	CAT_Co
	CAT_Cn
)

func EachGeneralCategory(yield func(GeneralCategory) bool) {
	for i := CAT_Lu; i <= CAT_Cn; i++ {
		if !yield(i) {
			break
		}
	}
}

var strToGeneralCategory map[string]GeneralCategory

func init() {
	strToGeneralCategory = make(map[string]GeneralCategory)
	for cat := range EachGeneralCategory {
		strToGeneralCategory[cat.String()] = cat
	}
}

func ParseGeneralCategory(s string) (GeneralCategory, error) {
	if c, ok := strToGeneralCategory[s]; ok {
		return c, nil
	}
	return GeneralCategory(0), fmt.Errorf("unknown general category: %s", s)
}

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type EastAsianWidth -trimprefix EAW_

type EastAsianWidth int

const (
	EAW_W EastAsianWidth = iota
	EAW_F
	EAW_A
	EAW_N
	EAW_Na
	EAW_H
)

func EachEastAsianWidth(yield func(EastAsianWidth) bool) {
	for i := EAW_W; i <= EAW_H; i++ {
		if !yield(i) {
			break
		}
	}
}

var strToEastAsianWidth map[string]EastAsianWidth

func init() {
	strToEastAsianWidth = make(map[string]EastAsianWidth)
	for eaw := range EachEastAsianWidth {
		strToEastAsianWidth[eaw.String()] = eaw
	}
}

func ParseEastAsianWidth(s string) (EastAsianWidth, error) {
	if c, ok := strToEastAsianWidth[s]; ok {
		return c, nil
	}
	return EastAsianWidth(0), fmt.Errorf("unknown east asian width: %s", s)
}
