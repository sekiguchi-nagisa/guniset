package op

type CaseFoldMap struct {
	fold   map[rune]rune
	unfold map[rune][]rune
}

func NewCaseFoldMap(runes [][2]rune) *CaseFoldMap {
	fold := make(map[rune]rune)
	unfold := make(map[rune][]rune)
	for _, r := range runes {
		fold[r[0]] = r[1]
		if _, ok := unfold[r[1]]; !ok {
			unfold[r[1]] = []rune{}
		}
		unfold[r[1]] = append(unfold[r[1]], r[0])
	}
	return &CaseFoldMap{fold: fold, unfold: unfold}
}

func (m *CaseFoldMap) LookupFold(r rune) rune {
	if a, ok := m.fold[r]; ok {
		return a
	}
	return r
}

func (m *CaseFoldMap) LookupUnfold(r rune) []rune {
	if a, ok := m.unfold[r]; ok {
		return a
	}
	return []rune{r}
}
