package op

import (
	"fmt"
	"slices"
	"strings"
)

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

func (c GeneralCategory) Format(aliasMap *AliasMap) string {
	abbr := c.String()
	tmp := []string{abbr}
	tmp = append(tmp, aliasMap.Lookup(abbr)...)
	return strings.Join(tmp, ", ")
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

func (e EastAsianWidth) Format(aliasMap *AliasMap) string {
	abbr := e.String()
	tmp := []string{abbr}
	tmp = append(tmp, aliasMap.Lookup(abbr)...)
	return strings.Join(tmp, ", ")
}

const ScriptPrefix = "sc"

func IsScriptPrefix(s string) bool {
	return s == ScriptPrefix
}

type Script int

type ScriptDef struct {
	scriptToAbbr []string
	abbrToScript map[string]Script
	unknown      Script
}

func NewScriptDef(longs []string, aliasMap *AliasMap) *ScriptDef {
	longs = append(longs, "Unknown")
	s := &ScriptDef{
		scriptToAbbr: make([]string, len(longs)),
		abbrToScript: make(map[string]Script),
		unknown:      Script(len(longs) - 1),
	}
	for i, long := range longs {
		abbr := aliasMap.LookupAbbr(long)
		s.scriptToAbbr[i] = abbr
		s.abbrToScript[abbr] = Script(i)
	}
	return s
}

func (d *ScriptDef) GetAbbr(s Script) string {
	return d.scriptToAbbr[s]
}

func (d *ScriptDef) EachScript(yield func(Script) bool) {
	for i := 0; i < len(d.scriptToAbbr); i++ {
		if !yield(Script(i)) {
			break
		}
	}
}

func (d *ScriptDef) Parse(s string, aliasMap *AliasMap) (Script, error) {
	if c, ok := d.abbrToScript[s]; ok {
		return c, nil
	}
	if aliasMap != nil {
		if abbr := aliasMap.LookupAbbr(s); len(abbr) > 0 {
			if c, ok := d.abbrToScript[abbr]; ok {
				return c, nil
			}
		}
	}
	return Script(0), fmt.Errorf("unknown script: %s", s)
}

func (d *ScriptDef) Unknown() Script {
	return d.unknown
}

func (d *ScriptDef) Format(s Script, aliasMap *AliasMap) string {
	abbr := d.GetAbbr(s)
	tmp := []string{abbr}
	tmp = append(tmp, aliasMap.Lookup(abbr)...)
	return strings.Join(tmp, ", ")
}

const ScriptExtensionPrefix = "scx"

func IsScriptExtensionPrefix(s string) bool {
	return s == ScriptExtensionPrefix
}

// PropertyDef common Unicode property definition
type PropertyDef[T ~int] struct {
	propertyToName []string
	nameToProperty map[string]T
}

func NewPropertyDef[T ~int](names []string) *PropertyDef[T] {
	ret := &PropertyDef[T]{
		propertyToName: slices.Clone(names),
		nameToProperty: make(map[string]T),
	}
	for i, name := range names {
		p := T(i)
		ret.nameToProperty[name] = p
	}
	return ret
}

func (d *PropertyDef[T]) Parse(s string) (T, error) {
	if p, ok := d.nameToProperty[s]; ok {
		return p, nil
	}
	return T(0), fmt.Errorf("unknown property: %s", s)
}

func (d *PropertyDef[T]) GetName(p T) string {
	return d.propertyToName[p]
}

func (d *PropertyDef[T]) EachProperty(yield func(T) bool) {
	for i := 0; i < len(d.propertyToName); i++ {
		if !yield(T(i)) {
			break
		}
	}
}

func (d *PropertyDef[T]) Format(p T) string {
	return d.GetName(p)
}

type PropList int

const PropListPrefix = "prop"

func IsPropListPrefix(s string) bool {
	return s == PropListPrefix
}

type DerivedCoreProperty int

const DerivedCorePropPrefix = "dcp"

func IsDerivedCorePropertyPrefix(s string) bool {
	return s == DerivedCorePropPrefix
}

type Emoji int

const EmojiPrefix = "emoji"

func IsEmojiPrefix(s string) bool {
	return s == EmojiPrefix
}

type DerivedBinaryProperty int

const DerivedBinaryPropPrefix = "dbp"

func IsDerivedBinaryPropertyPrefix(s string) bool {
	return s == DerivedBinaryPropPrefix
}

type DerivedNormalizationProp int

const DerivedNormalizationPropPrefix = "dnp"

func IsDerivedNormalizationPropPrefix(s string) bool {
	return s == DerivedNormalizationPropPrefix
}

type GraphemeBreakProperty int

const GraphemeBreakPropPrefix = "gbp"

func IsGraphemeBreakPropertyPrefix(s string) bool {
	return s == GraphemeBreakPropPrefix
}

type WordBreakProperty int

const WordBreakPropPrefix = "wbp"

func IsWordBreakPropertyPrefix(s string) bool {
	return s == WordBreakPropPrefix
}

type SentenceBreakProperty int

const SentenceBreakPropPrefix = "sbp"

func IsSentenceBreakPropertyPrefix(s string) bool {
	return s == SentenceBreakPropPrefix
}
