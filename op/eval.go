package op

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/sekiguchi-nagisa/guniset/set"
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

func (e *EvalContext) FillEawN() *set.UniSet {
	eawSet := e.eawSet[EAW_N]
	if eawSet != nil {
		return eawSet
	}
	tmpSet := set.NewUniSetAll()
	builder := set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw != EAW_N {
			builder.AddSet(e.eawSet[eaw])
		}
	}
	removing := builder.Build()
	tmpSet.RemoveSet(&removing)
	e.eawSet[EAW_N] = &tmpSet
	return e.eawSet[EAW_N]
}

func (e *EvalContext) Query(r rune, writer io.Writer) error {
	cat := CAT_Cn
	eaw := EAW_N
	for cc, uniSet := range e.catSet {
		if uniSet.Find(r) {
			cat = cc
			break
		}
	}
	for e, uniSet := range e.eawSet {
		if uniSet.Find(r) {
			eaw = e
			break
		}
	}
	_, err := fmt.Fprintf(writer, "CodePoint: U+%04X\n"+
		"GeneralCategory: %v\n"+
		"EastAsianWidth: %v\n", r, cat, eaw)
	return err
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
	builderMap := map[GeneralCategory]*set.UniSetBuilder{}
	for cate := range EachGeneralCategory {
		builderMap[cate] = &set.UniSetBuilder{}
	}
	lr := NewLineReader("DerivedGeneralCategory.txt", reader)
	for lr.next() {
		line := lr.line()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// extract runeRange
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
		builderMap[cate].AddRange(set.RuneRange{
			First: first,
			Last:  last,
		})
	}
	err := lr.err()
	if err != nil {
		return nil, lr.formatErr(err)
	}

	// build
	ret := map[GeneralCategory]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		ret[cate] = &tmp
	}
	return ret, nil
}

func LoadEastAsianWidthMap(reader io.Reader) (map[EastAsianWidth]*set.UniSet, error) {
	builderMap := map[EastAsianWidth]*set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw == EAW_N {
			continue // fill N later
		}
		builderMap[eaw] = &set.UniSetBuilder{}
	}
	lr := NewLineReader("EastAsianWidth.txt", reader)
	for lr.next() {
		line := lr.line()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// extract runeRange
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
		if eaw == EAW_N {
			continue // fill N later
		}
		builderMap[eaw].AddRange(set.RuneRange{
			First: first,
			Last:  last,
		})
	}
	err := lr.err()
	if err != nil {
		return nil, lr.formatErr(err)
	}

	// build
	ret := map[EastAsianWidth]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		ret[cate] = &tmp
	}
	return ret, nil
}
