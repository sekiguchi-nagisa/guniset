package op

import (
	"slices"
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

type StringPropertyMap = map[string][]String

func LookupStringPropertyValues(propertyMap StringPropertyMap, property string) []String {
	if property == "RGI_Emoji" {
		var list []String
		for _, v := range propertyMap {
			list = append(list, v...)
		}
		slices.SortFunc(list, func(a, b String) int {
			return strings.Compare(a.String(), b.String())
		})
		return list
	}
	return propertyMap[property]
}

func Properties(propertyMap StringPropertyMap) []string {
	var list []string
	for k := range propertyMap {
		list = append(list, k)
	}
	list = append(list, "RGI_Emoji")
	slices.Sort(list)
	return list
}
