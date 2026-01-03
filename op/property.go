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

var abbrToGeneralCategory map[string]GeneralCategory

var longToGeneralCategory map[string]GeneralCategory

func init() {
	abbrToGeneralCategory = make(map[string]GeneralCategory)
	for cat := range EachGeneralCategoryAll {
		abbrToGeneralCategory[cat.String()] = cat
	}

	// define long name
	longToGeneralCategory = make(map[string]GeneralCategory)
	// L
	longToGeneralCategory["Uppercase_Letter"] = CAT_Lu
	longToGeneralCategory["Lowercase_Letter"] = CAT_Ll
	longToGeneralCategory["Titlecase_Letter"] = CAT_Lt
	longToGeneralCategory["Cased_Letter"] = CAT_LC
	longToGeneralCategory["Modifier_Letter"] = CAT_Lm
	longToGeneralCategory["Other_Letter"] = CAT_Lo
	longToGeneralCategory["Letter"] = CAT_L
	// M
	longToGeneralCategory["Nonspacing_Mark"] = CAT_Mn
	longToGeneralCategory["Spacing_Mark"] = CAT_Mc
	longToGeneralCategory["Enclosing_Mark"] = CAT_Me
	longToGeneralCategory["Mark"] = CAT_M
	// N
	longToGeneralCategory["Decimal_Number"] = CAT_Nd
	longToGeneralCategory["Letter_Number"] = CAT_Nl
	longToGeneralCategory["Other_Number"] = CAT_No
	longToGeneralCategory["Number"] = CAT_N
	// P
	longToGeneralCategory["Connector_Punctuation"] = CAT_Pc
	longToGeneralCategory["Dash_Punctuation"] = CAT_Pd
	longToGeneralCategory["Open_Punctuation"] = CAT_Ps
	longToGeneralCategory["Close_Punctuation"] = CAT_Pe
	longToGeneralCategory["Initial_Punctuation"] = CAT_Pi
	longToGeneralCategory["Final_Punctuation"] = CAT_Pf
	longToGeneralCategory["Other_Punctuation"] = CAT_Po
	longToGeneralCategory["Punctuation"] = CAT_P
	// S
	longToGeneralCategory["Math_Symbol"] = CAT_Sm
	longToGeneralCategory["Currency_Symbol"] = CAT_Sc
	longToGeneralCategory["Modifier_Symbol"] = CAT_Sk
	longToGeneralCategory["Other_Symbol"] = CAT_So
	longToGeneralCategory["Symbol"] = CAT_S
	// Z
	longToGeneralCategory["Space_Separator"] = CAT_Zs
	longToGeneralCategory["Line_Separator"] = CAT_Zl
	longToGeneralCategory["Paragraph_Separator"] = CAT_Zp
	longToGeneralCategory["Separator"] = CAT_Z
	// C
	longToGeneralCategory["Control"] = CAT_Cc
	longToGeneralCategory["Format"] = CAT_Cf
	longToGeneralCategory["Surrogate"] = CAT_Cs
	longToGeneralCategory["Private_Use"] = CAT_Co
	longToGeneralCategory["Unassigned"] = CAT_Cn
	longToGeneralCategory["Other"] = CAT_C
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
