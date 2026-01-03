package op

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/sekiguchi-nagisa/guniset/set"
)

type DataHeader struct {
	Filename string
	Created  string
}

type DataHeaders struct {
	List []DataHeader
}

func (d *DataHeaders) Print(writer io.Writer) error {
	for _, info := range d.List {
		_, err := fmt.Fprintf(writer, "%s\n%s\n", info.Filename, info.Created)
		if err != nil {
			return err
		}
	}
	return nil
}

type UniSetMap[T comparable] = map[T]*set.UniSet

type EvalContext struct {
	Headers   DataHeaders
	CateMap   UniSetMap[GeneralCategory]
	EawMap    UniSetMap[EastAsianWidth]
	AliasMaps AliasMaps
}

func NewEvalContext(unicodeData io.ReadCloser, eastAsianWidth io.ReadCloser, aliases io.ReadCloser) (*EvalContext, error) {
	headers := DataHeaders{}
	catMap, err := LoadGeneralCategoryMap(unicodeData, &headers)
	if err != nil {
		return nil, err
	}
	eawMap, err := LoadEastAsianWidthMap(eastAsianWidth, &headers)
	if err != nil {
		return nil, err
	}
	aliasMaps, err := LoadTargetAliasMap(aliases, &headers,
		map[string]struct{}{GeneralCategoryPrefix: {}, EastAsianWidthPrefix: {}})
	if err != nil {
		return nil, err
	}
	return &EvalContext{
		Headers:   headers,
		CateMap:   catMap,
		EawMap:    eawMap,
		AliasMaps: aliasMaps,
	}, nil
}

func (e *EvalContext) FillEawN() *set.UniSet {
	eawSet := e.EawMap[EAW_N]
	if eawSet != nil {
		return eawSet
	}
	tmpSet := set.NewUniSetAll()
	builder := set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw != EAW_N {
			builder.AddSet(e.EawMap[eaw])
		}
	}
	removing := builder.Build()
	tmpSet.RemoveSet(&removing)
	e.EawMap[EAW_N] = &tmpSet
	return e.EawMap[EAW_N]
}

func (e *EvalContext) Query(r rune, writer io.Writer) error {
	cat := CAT_Cn
	eaw := EAW_N
	for cc, uniSet := range e.CateMap {
		if uniSet.Find(r) {
			cat = cc
			break
		}
	}
	for e, uniSet := range e.EawMap {
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

type DataLoader struct {
	name    string
	scanner *bufio.Scanner
	lineno  int
	header  DataHeader
}

func NewDataLoader(name string, reader io.Reader) DataLoader {
	return DataLoader{name: name, scanner: bufio.NewScanner(reader), lineno: 0}
}

func (d *DataLoader) next() bool {
	ok := d.scanner.Scan()
	if ok {
		d.lineno++
	}
	return ok
}

func (d *DataLoader) line() string {
	return d.scanner.Text()
}

func (d *DataLoader) err() error {
	return d.scanner.Err()
}

func (d *DataLoader) formatErr(e error) error {
	return fmt.Errorf("%s:%d: [load error] %s", d.name, d.lineno, e.Error())
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

func (d *DataLoader) Load(reader io.ReadCloser, callback func(string) error) error {
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	for d.next() {
		line := d.line()
		if d.lineno == 1 && strings.HasPrefix(line, "#") {
			d.header.Filename = strings.TrimPrefix(line, "# ")
			continue
		}
		if d.lineno == 2 && strings.HasPrefix(line, "#") {
			d.header.Created = strings.TrimPrefix(line, "# ")
			continue
		}
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		// parse entry
		err := callback(line)
		if err != nil {
			return d.formatErr(err)
		}
	}
	err := d.err()
	if err != nil {
		return d.formatErr(err)
	}
	return nil
}

func (d *DataLoader) LoadProperties(reader io.ReadCloser, callback func(set.RuneRange, string) error) error {
	return d.Load(reader, func(line string) error {
		runeRange, property, err := parseEntry(line)
		if err != nil {
			return err
		}
		return callback(runeRange, property)
	})
}

func LoadGeneralCategoryMap(reader io.ReadCloser, dbInfoList *DataHeaders) (setMap UniSetMap[GeneralCategory], e error) {
	builderMap := map[GeneralCategory]*set.UniSetBuilder{}
	for cate := range EachGeneralCategory {
		builderMap[cate] = &set.UniSetBuilder{}
	}

	// load
	loader := NewDataLoader("DerivedGeneralCategory.txt", reader)
	err := loader.LoadProperties(reader, func(runeRange set.RuneRange, property string) error {
		cate, err := ParseGeneralCategory(property, nil)
		if err != nil {
			return err
		}
		builderMap[cate].AddRange(runeRange)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// build
	setMap = map[GeneralCategory]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap[cate] = &tmp
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return
}

func LoadEastAsianWidthMap(reader io.ReadCloser, dbInfoList *DataHeaders) (setMap UniSetMap[EastAsianWidth], e error) {
	builderMap := map[EastAsianWidth]*set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw == EAW_N {
			continue // fill N later
		}
		builderMap[eaw] = &set.UniSetBuilder{}
	}

	// load
	loader := NewDataLoader("EastAsianWidth.txt", reader)
	err := loader.LoadProperties(reader, func(runeRange set.RuneRange, property string) error {
		eaw, err := ParseEastAsianWidth(property, nil)
		if err != nil {
			return err
		}
		if eaw != EAW_N { // skip N (fill later)
			builderMap[eaw].AddRange(runeRange)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// build
	setMap = map[EastAsianWidth]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap[cate] = &tmp
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return
}

func LoadTargetAliasMap(reader io.ReadCloser, dbInfoList *DataHeaders, targets map[string]struct{}) (aliasMaps AliasMaps, e error) {
	aliasMaps = AliasMaps{}
	for target := range targets {
		aliasMaps[target] = NewAliasMap(target)
	}
	loader := NewDataLoader("PropertyValueAliases.txt", reader)
	err := loader.Load(reader, func(line string) error {
		if ret, ok := ParseAliasEntry(line, targets); ok {
			aliasMaps[ret.property].AddAll(ret.abbr, ret.longs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return aliasMaps, nil
}
