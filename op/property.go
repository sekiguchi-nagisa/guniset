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

var longNameSet = map[GeneralCategory]string{
	CAT_Lu: "Uppercase_Letter",
	CAT_Ll: "Lowercase_Letter",
	CAT_Lt: "Titlecase_Letter",
	CAT_LC: "Cased_Letter",
	CAT_Lm: "Modifier_Letter",
	CAT_Lo: "Other_Letter",
	CAT_L:  "Letter",
	CAT_Mn: "Nonspacing_Mark",
	CAT_Mc: "Spacing_Mark",
	CAT_Me: "Enclosing_Mark",
	CAT_M:  "Mark",
	CAT_Nd: "Decimal_Number",
	CAT_Nl: "Letter_Number",
	CAT_No: "Other_Number",
	CAT_N:  "Number",
	CAT_Pc: "Connector_Punctuation",
	CAT_Pd: "Dash_Punctuation",
	CAT_Ps: "Open_Punctuation",
	CAT_Pe: "Close_Punctuation",
	CAT_Pi: "Initial_Punctuation",
	CAT_Pf: "Final_Punctuation",
	CAT_Po: "Other_Punctuation",
	CAT_P:  "Punctuation",
	CAT_Sm: "Math_Symbol",
	CAT_Sc: "Currency_Symbol",
	CAT_Sk: "Modifier_Symbol",
	CAT_So: "Other_Symbol",
	CAT_S:  "Symbol",
	CAT_Zs: "Space_Separator",
	CAT_Zl: "Line_Separator",
	CAT_Zp: "Paragraph_Separator",
	CAT_Z:  "Separator",
	CAT_Cc: "Control",
	CAT_Cf: "Format",
	CAT_Cs: "Surrogate",
	CAT_Co: "Private_Use",
	CAT_Cn: "Unassigned",
	CAT_C:  "Other",
}

func (c GeneralCategory) LongName() string {
	return longNameSet[c]
}

var abbrToGeneralCategory map[string]GeneralCategory

var longToGeneralCategory map[string]GeneralCategory

func init() {
	abbrToGeneralCategory = make(map[string]GeneralCategory)
	for cat := range EachGeneralCategoryAll {
		abbrToGeneralCategory[cat.String()] = cat
	}
	longToGeneralCategory = make(map[string]GeneralCategory)
	for cat := range EachGeneralCategoryAll {
		longToGeneralCategory[cat.LongName()] = cat
	}
}

func ParseGeneralCategory(s string) (GeneralCategory, error) {
	if c, ok := abbrToGeneralCategory[s]; ok {
		return c, nil
	}
	if c, ok := longToGeneralCategory[s]; ok {
		return c, nil
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
