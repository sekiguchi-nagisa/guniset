package op

import (
	"bufio"
	"fmt"
	"github.com/sekiguchi-nagisa/guniset/set"
	"io"
	"strings"
)

type EvalContext struct {
	catSet map[GeneralCategory]*set.UniSet
	eawSet map[EastAsianWidth]*set.UniSet
}

func NewEvalContext(unicodeData io.Reader, eastAsianWidth io.Reader) (*EvalContext, error) {
	catSet, err := LoadGeneralCategoryMap(unicodeData)
	if err != nil {
		return nil, err
	}
	eawSet, err := LoadEastAsianWidthMap(eastAsianWidth)
	if err != nil {
		return nil, err
	}
	return &EvalContext{
		catSet: catSet,
		eawSet: eawSet,
	}, nil
}

func loadErr(loc string, lineno int, e error) error {
	return fmt.Errorf("%s:%d: [load error] %s", loc, lineno, e.Error())
}

func LoadGeneralCategoryMap(reader io.Reader) (map[GeneralCategory]*set.UniSet, error) {
	ret := map[GeneralCategory]*set.UniSet{}
	for cate := range EachGeneralCategory {
		ret[cate] = &set.UniSet{}
	}
	loc := "UnicodeData.txt"
	scanner := bufio.NewScanner(reader)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++
		ss := strings.Split(line, ";")
		r, err := set.ParseRune(ss[0])
		if err != nil {
			return nil, loadErr(loc, lineno, err)
		}
		cate, err := ParseGeneralCategory(ss[2])
		if err != nil {
			return nil, loadErr(loc, lineno, err)
		}
		if !ret[cate].Add(r) {
			return nil, loadErr(loc, lineno,
				fmt.Errorf("rune %04x already found in %s", r, cate.String()))
		}
	}
	err := scanner.Err()
	if err != nil {
		return nil, loadErr(loc, lineno, err)
	}
	return ret, nil
}

func LoadEastAsianWidthMap(reader io.Reader) (map[EastAsianWidth]*set.UniSet, error) {
	ret := map[EastAsianWidth]*set.UniSet{}
	for eaw := range EachEastAsianWidth {
		ret[eaw] = &set.UniSet{}
	}
	loc := "EastAsianWidth.txt"
	scanner := bufio.NewScanner(reader)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// extract interval
		ss := strings.Split(line, ";")
		cc := strings.Split(strings.TrimSpace(ss[0]), "..")
		first, err := set.ParseRune(cc[0])
		if err != nil {
			return nil, loadErr(loc, lineno, err)
		}
		last := first
		if len(cc) == 2 {
			last, err = set.ParseRune(cc[1])
			if err != nil {
				return nil, loadErr(loc, lineno, err)
			}
		}

		// extract EAW
		v := strings.TrimSpace(strings.Split(strings.TrimSpace(ss[1]), "#")[0])
		eaw, err := ParseEastAsianWidth(v)
		if err != nil {
			return nil, loadErr(loc, lineno, err)
		}
		ret[eaw].AddInterval(set.RuneInterval{
			First: first,
			Last:  last,
		})
	}
	err := scanner.Err()
	if err != nil {
		return nil, loadErr(loc, lineno, err)
	}
	return ret, nil
}
