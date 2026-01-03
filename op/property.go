package op

import "fmt"

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type GeneralCategory -trimprefix CAT_

type GeneralCategory int

const GeneralCategoryPrefix = "gc"

var generalCategoryPrefixes = []string{
	"cat", "gc",
}

func IsGeneralCategoryPrefix(s string) bool {
	for _, prefix := range generalCategoryPrefixes {
		if s == prefix {
			return true
		}
	}
	return false
}

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
	CAT_LC // Lu | Ll | Lt
	CAT_L  // Lu | Ll | Lt | Lm | Lo
	CAT_M  // Mn | Mc | Me
	CAT_N  // Nd | Nl | No
	CAT_P  // Pc | Pd | Ps | Pe | Pi | Pf | Po
	CAT_S  // Sm | Sc | Sk | So
	CAT_Z  // Zs | Zl | Zp
	CAT_C  // Cc | Cf | Cs | Co | Cn
)

func EachGeneralCategory(yield func(GeneralCategory) bool) {
	for i := CAT_Lu; i <= CAT_Cn; i++ {
		if !yield(i) {
			break
		}
	}
}

func EachGeneralCategoryAll(yield func(GeneralCategory) bool) {
	for i := CAT_Lu; i <= CAT_C; i++ {
		if !yield(i) {
			break
		}
	}
}

var abbrToGeneralCategory map[string]GeneralCategory

func init() {
	abbrToGeneralCategory = make(map[string]GeneralCategory)
	for cat := range EachGeneralCategoryAll {
		abbrToGeneralCategory[cat.String()] = cat
	}
}

func ParseGeneralCategory(s string, aliasMap *AliasMap) (GeneralCategory, error) {
	if c, ok := abbrToGeneralCategory[s]; ok {
		return c, nil
	}
	if aliasMap != nil {
		if abbr := aliasMap.LookupAbbr(s); len(abbr) > 0 {
			if c, ok := abbrToGeneralCategory[abbr]; ok {
				return c, nil
			}
		}
	}
	return GeneralCategory(0), fmt.Errorf("unknown general category: %s", s)
}

var combinedGeneralCategory = map[GeneralCategory][]GeneralCategory{
	CAT_LC: {CAT_Lu, CAT_Ll, CAT_Lt},
	CAT_L:  {CAT_Lu, CAT_Ll, CAT_Lt, CAT_Lm, CAT_Lo},
	CAT_M:  {CAT_Mn, CAT_Mc, CAT_Me},
	CAT_N:  {CAT_Nd, CAT_Nl, CAT_No},
	CAT_P:  {CAT_Pc, CAT_Pd, CAT_Ps, CAT_Pe, CAT_Pi, CAT_Pf, CAT_Po},
	CAT_S:  {CAT_Sm, CAT_Sc, CAT_Sk, CAT_So},
	CAT_Z:  {CAT_Zs, CAT_Zl, CAT_Zp},
	CAT_C:  {CAT_Cc, CAT_Cf, CAT_Cs, CAT_Co, CAT_Cn},
}

func (c GeneralCategory) Combinations() []GeneralCategory {
	return combinedGeneralCategory[c]
}

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type EastAsianWidth -trimprefix EAW_

type EastAsianWidth int

const EastAsianWidthPrefix = "ea"

var eastAsianWidthPrefixes = []string{
	"eaw", "ea",
}

func IsEastAsianWidthPrefix(s string) bool {
	for _, prefix := range eastAsianWidthPrefixes {
		if s == prefix {
			return true
		}
	}
	return false
}

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

var abbrToEastAsianWidth map[string]EastAsianWidth

func init() {
	abbrToEastAsianWidth = make(map[string]EastAsianWidth)
	for eaw := range EachEastAsianWidth {
		abbrToEastAsianWidth[eaw.String()] = eaw
	}
}

func ParseEastAsianWidth(s string, aliasMap *AliasMap) (EastAsianWidth, error) {
	if c, ok := abbrToEastAsianWidth[s]; ok {
		return c, nil
	}
	if aliasMap != nil {
		if abbr := aliasMap.LookupAbbr(s); len(abbr) > 0 {
			if c, ok := abbrToEastAsianWidth[abbr]; ok {
				return c, nil
			}
		}
	}
	return EastAsianWidth(0), fmt.Errorf("unknown east asian width: %s", s)
}
