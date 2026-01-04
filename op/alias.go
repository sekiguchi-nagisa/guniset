package op

import (
	"fmt"
	"strings"
)

type AliasMap struct {
	abbrToLong map[string][]string
	longToAbbr map[string]string
}

func NewAliasMap() *AliasMap {
	return &AliasMap{
		abbrToLong: make(map[string][]string),
		longToAbbr: make(map[string]string),
	}
}

func (a *AliasMap) Add(abbr string, long string) {
	a.abbrToLong[abbr] = append(a.abbrToLong[abbr], long)
	a.longToAbbr[long] = abbr
}

func (a *AliasMap) AddAll(abbr string, longs []string) {
	a.abbrToLong[abbr] = append(a.abbrToLong[abbr], longs...)
	for _, long := range longs {
		a.longToAbbr[long] = abbr
	}
}

func (a *AliasMap) Lookup(abbr string) []string {
	return a.abbrToLong[abbr]
}

func (a *AliasMap) LookupAbbr(long string) string {
	return a.longToAbbr[long]
}

type AliasMaps = map[string]*AliasMap

type AliasMapRecord struct {
	gc *AliasMap
	ea *AliasMap
	sc *AliasMap
}

func NewAliasMapRecord() *AliasMapRecord {
	return &AliasMapRecord{
		gc: NewAliasMap(),
		ea: NewAliasMap(),
		sc: NewAliasMap(),
	}
}

func (a *AliasMapRecord) Category() *AliasMap {
	return a.gc
}

func (a *AliasMapRecord) Eaw() *AliasMap {
	return a.ea
}

func (a *AliasMapRecord) Script() *AliasMap {
	return a.sc
}

var aliasTargetPrefixes = map[string]struct{}{
	GeneralCategoryPrefix: {},
	EastAsianWidthPrefix:  {},
	ScriptPrefix:          {},
}

func ParseAliasEntry(line string) (struct {
	property string
	abbr     string
	longs    []string
}, bool) {
	ret := struct {
		property string
		abbr     string
		longs    []string
	}{}
	line = strings.Split(line, "#")[0] // trim comment
	ss := strings.Split(line, ";")
	ret.property = strings.TrimSpace(ss[0])
	if _, ok := aliasTargetPrefixes[ret.property]; !ok {
		return ret, false
	}
	ret.abbr = strings.TrimSpace(ss[1])
	ret.longs = make([]string, 0, len(ss)-2)
	for _, s := range ss[2:] {
		ret.longs = append(ret.longs, strings.TrimSpace(s))
	}
	return ret, true
}

func (a *AliasMapRecord) Define(prefix string, abbr string, longs []string) error {
	switch prefix {
	case GeneralCategoryPrefix:
		a.gc.AddAll(abbr, longs)
	case EastAsianWidthPrefix:
		a.ea.AddAll(abbr, longs)
	case ScriptPrefix:
		a.sc.AddAll(abbr, longs)
	default:
		return fmt.Errorf("unknown prefix: %s", prefix)
	}
	return nil
}

func (a *AliasMapRecord) Resolve(line string) error {
	if ret, ok := ParseAliasEntry(line); ok {
		return a.Define(ret.property, ret.abbr, ret.longs)
	}
	return nil
}
