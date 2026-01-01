package op

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/sekiguchi-nagisa/guniset/set"
)

type UniSetMap[T comparable] struct {
	// extracted from a database header
	Filename string
	Created  string

	// generated from actual database content
	Map map[T]*set.UniSet
}

func (t *UniSetMap[T]) PrintHeader(writer io.Writer) error {
	_, err := writer.Write([]byte(fmt.Sprintf("%s\n%s\n", t.Filename, t.Created)))
	return err
}

type EvalContext struct {
	CateMap UniSetMap[GeneralCategory]
	EawMap  UniSetMap[EastAsianWidth]
}

func NewEvalContext(unicodeData io.Reader, eastAsianWidth io.Reader) (*EvalContext, error) {
	catMap, err := LoadGeneralCategoryMap(unicodeData)
	if err != nil {
		return nil, err
	}
	eawMap, err := LoadEastAsianWidthMap(eastAsianWidth)
	if err != nil {
		return nil, err
	}
	return &EvalContext{
		CateMap: catMap,
		EawMap:  eawMap,
	}, nil
}

func (e *EvalContext) FillEawN() *set.UniSet {
	eawSet := e.EawMap.Map[EAW_N]
	if eawSet != nil {
		return eawSet
	}
	tmpSet := set.NewUniSetAll()
	builder := set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw != EAW_N {
			builder.AddSet(e.EawMap.Map[eaw])
		}
	}
	removing := builder.Build()
	tmpSet.RemoveSet(&removing)
	e.EawMap.Map[EAW_N] = &tmpSet
	return e.EawMap.Map[EAW_N]
}

func (e *EvalContext) Query(r rune, writer io.Writer) error {
	cat := CAT_Cn
	eaw := EAW_N
	for cc, uniSet := range e.CateMap.Map {
		if uniSet.Find(r) {
			cat = cc
			break
		}
	}
	for e, uniSet := range e.EawMap.Map {
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

func parseEntry(line string) (runeRange set.RuneRange, property string, err error) {
	// extract runeRange
	ss := strings.Split(line, ";")
	cc := strings.Split(strings.TrimSpace(ss[0]), "..")
	runeRange.First, err = set.ParseRune(cc[0])
	if err != nil {
		return
	}
	runeRange.Last = runeRange.First
	if len(cc) == 2 {
		runeRange.Last, err = set.ParseRune(cc[1])
		if err != nil {
			return
		}
	}

	// extract property
	property = strings.TrimSpace(strings.Split(strings.TrimSpace(ss[1]), "#")[0])
	return
}

func LoadGeneralCategoryMap(reader io.Reader) (setMap UniSetMap[GeneralCategory], e error) {
	builderMap := map[GeneralCategory]*set.UniSetBuilder{}
	for cate := range EachGeneralCategory {
		builderMap[cate] = &set.UniSetBuilder{}
	}
	lr := NewLineReader("DerivedGeneralCategory.txt", reader)
	for lr.next() {
		line := lr.line()
		if lr.lineno == 1 && strings.HasPrefix(line, "#") {
			setMap.Filename = strings.TrimPrefix(line, "# ")
			continue
		}
		if lr.lineno == 2 && strings.HasPrefix(line, "#") {
			setMap.Created = strings.TrimPrefix(line, "# ")
			continue
		}
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// parse entry
		runeRange, property, err := parseEntry(line)
		if err != nil {
			e = lr.formatErr(err)
			return
		}
		cate, err := ParseGeneralCategory(property)
		if err != nil {
			e = lr.formatErr(err)
			return
		}
		builderMap[cate].AddRange(runeRange)
	}
	err := lr.err()
	if err != nil {
		e = lr.formatErr(err)
		return
	}

	// build
	setMap.Map = map[GeneralCategory]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap.Map[cate] = &tmp
	}
	return
}

func LoadEastAsianWidthMap(reader io.Reader) (setMap UniSetMap[EastAsianWidth], e error) {
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
		if lr.lineno == 1 && strings.HasPrefix(line, "#") {
			setMap.Filename = strings.TrimPrefix(line, "# ")
			continue
		}
		if lr.lineno == 2 && strings.HasPrefix(line, "#") {
			setMap.Created = strings.TrimPrefix(line, "# ")
			continue
		}
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// parse entry
		runeRange, property, err := parseEntry(line)
		if err != nil {
			e = lr.formatErr(err)
			return
		}
		eaw, err := ParseEastAsianWidth(property)
		if err != nil {
			e = lr.formatErr(err)
			return
		}
		if eaw == EAW_N {
			continue // skip (fill N later)
		}
		builderMap[eaw].AddRange(runeRange)
	}
	err := lr.err()
	if err != nil {
		e = lr.formatErr(err)
		return
	}

	// build
	setMap.Map = map[EastAsianWidth]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap.Map[cate] = &tmp
	}
	return
}
