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

type LineReader struct {
	name    string
	scanner *bufio.Scanner
	lineno  int
}

func NewLineReader(name string, reader io.Reader) LineReader {
	return LineReader{name: name, scanner: bufio.NewScanner(reader), lineno: 0}
}

func (lr *LineReader) next() bool {
	ok := lr.scanner.Scan()
	if ok {
		lr.lineno++
	}
	return ok
}

func (lr *LineReader) line() string {
	return lr.scanner.Text()
}

func (lr *LineReader) err() error {
	return lr.scanner.Err()
}

func (lr *LineReader) formatErr(e error) error {
	return fmt.Errorf("%s:%d: [load error] %s", lr.name, lr.lineno, e.Error())
}

func LoadGeneralCategoryMap(reader io.Reader) (map[GeneralCategory]*set.UniSet, error) {
	ret := map[GeneralCategory]*set.UniSet{}
	for cate := range EachGeneralCategory {
		ret[cate] = &set.UniSet{}
	}
	lr := NewLineReader("DerivedGeneralCategory.txt", reader)
	for lr.next() {
		line := lr.line()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// extract interval
		ss := strings.Split(line, ";")
		cc := strings.Split(strings.TrimSpace(ss[0]), "..")
		first, err := set.ParseRune(cc[0])
		if err != nil {
			return nil, lr.formatErr(err)
		}
		last := first
		if len(cc) == 2 {
			last, err = set.ParseRune(cc[1])
			if err != nil {
				return nil, lr.formatErr(err)
			}
		}

		// extract EAW
		v := strings.TrimSpace(strings.Split(strings.TrimSpace(ss[1]), "#")[0])
		cate, err := ParseGeneralCategory(v)
		if err != nil {
			return nil, lr.formatErr(err)
		}
		ret[cate].AddInterval(set.RuneInterval{
			First: first,
			Last:  last,
		})
	}
	err := lr.err()
	if err != nil {
		return nil, lr.formatErr(err)
	}
	return ret, nil
}

func LoadEastAsianWidthMap(reader io.Reader) (map[EastAsianWidth]*set.UniSet, error) {
	ret := map[EastAsianWidth]*set.UniSet{}
	for eaw := range EachEastAsianWidth {
		ret[eaw] = &set.UniSet{}
	}
	lr := NewLineReader("EastAsianWidth.txt", reader)
	for lr.next() {
		line := lr.line()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// extract interval
		ss := strings.Split(line, ";")
		cc := strings.Split(strings.TrimSpace(ss[0]), "..")
		first, err := set.ParseRune(cc[0])
		if err != nil {
			return nil, lr.formatErr(err)
		}
		last := first
		if len(cc) == 2 {
			last, err = set.ParseRune(cc[1])
			if err != nil {
				return nil, lr.formatErr(err)
			}
		}

		// extract EAW
		v := strings.TrimSpace(strings.Split(strings.TrimSpace(ss[1]), "#")[0])
		eaw, err := ParseEastAsianWidth(v)
		if err != nil {
			return nil, lr.formatErr(err)
		}
		ret[eaw].AddInterval(set.RuneInterval{
			First: first,
			Last:  last,
		})
	}
	err := lr.err()
	if err != nil {
		return nil, lr.formatErr(err)
	}
	return ret, nil
}
