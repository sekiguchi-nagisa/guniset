package op

import "strings"

type AliasMap struct {
	name       string
	abbrToLong map[string][]string
	longToAbbr map[string]string
}

func NewAliasMap(name string) *AliasMap {
	return &AliasMap{
		name:       name,
		abbrToLong: make(map[string][]string),
		longToAbbr: make(map[string]string),
	}
}

func (a *AliasMap) Name() string {
	return a.name
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

func ParseAliasEntry(line string, targets map[string]struct{}) (struct {
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
	if _, ok := targets[ret.property]; !ok {
		return ret, false
	}
	ret.abbr = strings.TrimSpace(ss[1])
	ret.longs = make([]string, 0, len(ss)-2)
	for _, s := range ss[2:] {
		ret.longs = append(ret.longs, strings.TrimSpace(s))
	}
	return ret, true
}

type AliasMaps = map[string]*AliasMap
