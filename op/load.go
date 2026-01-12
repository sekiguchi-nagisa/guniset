package op

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
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
		_, err := fmt.Fprintf(writer, "- %s\n  %s\n", info.Filename, info.Created)
		if err != nil {
			return err
		}
	}
	return nil
}

type UnicodeData struct {
	GeneralCategory      string // DerivedGeneralCategory.txt
	EastAsianWidth       string // EastAsianWidth.txt
	Scripts              string // Scripts.txt
	ScriptExtensions     string // ScriptExtensions.txt
	PropertyValueAliases string // PropertyValueAliases.txt
	PropList             string // PropList.txt
}

func NewUnicodeData(unicodeDir string) *UnicodeData {
	return &UnicodeData{
		GeneralCategory:      path.Join(unicodeDir, "DerivedGeneralCategory.txt"),
		EastAsianWidth:       path.Join(unicodeDir, "EastAsianWidth.txt"),
		Scripts:              path.Join(unicodeDir, "Scripts.txt"),
		ScriptExtensions:     path.Join(unicodeDir, "ScriptExtensions.txt"),
		PropertyValueAliases: path.Join(unicodeDir, "PropertyValueAliases.txt"),
		PropList:             path.Join(unicodeDir, "PropList.txt"),
	}
}

type UniSetMap[T comparable] = map[T]*set.UniSet

type EvalContext struct {
	Headers        DataHeaders
	CateMap        UniSetMap[GeneralCategory]
	EawMap         UniSetMap[EastAsianWidth]
	AliasMapRecord *AliasMapRecord
	ScriptDef      *ScriptDef
	ScriptMap      UniSetMap[Script]
	ScriptXMap     UniSetMap[Script]
	PropListDef    *PropertyDef[PropList]
	PropListMap    UniSetMap[PropList]
}

func NewEvalContext(data *UnicodeData) (*EvalContext, error) {
	headers := DataHeaders{}
	catMap, err := LoadGeneralCategoryMap(data.GeneralCategory, &headers)
	if err != nil {
		return nil, err
	}
	eawMap, err := LoadEastAsianWidthMap(data.EastAsianWidth, &headers)
	if err != nil {
		return nil, err
	}
	aliasMaps, err := LoadTargetAliasMap(data.PropertyValueAliases, &headers)
	if err != nil {
		return nil, err
	}
	scriptDef, scriptMap, err := LoadScriptMap(data.Scripts, aliasMaps.Script(), &headers)
	if err != nil {
		return nil, err
	}
	scriptXMap, err := LoadScriptXMap(data.ScriptExtensions, scriptDef, aliasMaps.Script(), &headers)
	if err != nil {
		return nil, err
	}
	propDef, propListMap, err := LoadPropListMap(data.PropList, &headers)
	if err != nil {
		return nil, err
	}
	return &EvalContext{
		Headers:        headers,
		CateMap:        catMap,
		EawMap:         eawMap,
		AliasMapRecord: aliasMaps,
		ScriptDef:      scriptDef,
		ScriptMap:      scriptMap,
		ScriptXMap:     scriptXMap,
		PropListDef:    propDef,
		PropListMap:    propListMap,
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

func (e *EvalContext) FillScriptUnknown() *set.UniSet {
	scriptSet := e.ScriptMap[e.ScriptDef.Unknown()]
	if scriptSet != nil {
		return scriptSet
	}
	tmpSet := set.NewUniSetAll()
	builder := set.UniSetBuilder{}
	for sc := range e.ScriptDef.EachScript {
		if sc != e.ScriptDef.Unknown() {
			builder.AddSet(e.ScriptMap[sc])
		}
	}
	removing := builder.Build()
	tmpSet.RemoveSet(&removing)
	e.ScriptMap[e.ScriptDef.Unknown()] = &tmpSet
	return e.ScriptMap[e.ScriptDef.Unknown()]
}

func formatScriptX(def *ScriptDef, scx []Script) string {
	builder := strings.Builder{}
	builder.WriteString("[")
	for i, s := range scx {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(def.GetAbbr(s))
	}
	builder.WriteString("]")
	return builder.String()
}

func (e *EvalContext) Query(r rune, writer io.Writer) error {
	cat := CAT_Cn
	eaw := EAW_N
	sc := e.ScriptDef.Unknown()
	var scx []Script
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
	for s, uniSet := range e.ScriptMap {
		if uniSet.Find(r) {
			sc = s
			break
		}
	}
	for s, uniSet := range e.ScriptXMap {
		if uniSet.Find(r) {
			scx = append(scx, s)
		}
	}
	_, err := fmt.Fprintf(writer, "CodePoint: U+%04X\n"+
		"GeneralCategory: %s\nEastAsianWidth: %s\nScript: %s\nScriptExtension: %s\n", r,
		cat.Format(e.AliasMapRecord.Category()),
		eaw.Format(e.AliasMapRecord.Eaw()),
		e.ScriptDef.Format(sc, e.AliasMapRecord.Script()),
		formatScriptX(e.ScriptDef, scx))
	return err
}

type DataLoader struct {
	name    string
	file    *os.File
	scanner *bufio.Scanner
	lineno  int
	header  DataHeader
}

func NewDataLoader(p string) (DataLoader, error) {
	f, err := os.Open(p)
	if err != nil {
		return DataLoader{}, err
	}
	return DataLoader{name: path.Base(p), file: f, scanner: bufio.NewScanner(f), lineno: 0}, nil
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

func (d *DataLoader) Next(yield func(int, string) bool) {
	for d.next() {
		if !yield(d.lineno, d.line()) {
			break
		}
	}
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

func (d *DataLoader) Load(callback func(string) error) error {
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(d.file)
	for lineno, line := range d.Next {
		if lineno == 1 && strings.HasPrefix(line, "#") {
			d.header.Filename = strings.TrimPrefix(line, "# ")
			continue
		}
		if lineno == 2 && strings.HasPrefix(line, "#") {
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

func (d *DataLoader) LoadProperties(callback func(set.RuneRange, string) error) error {
	return d.Load(func(line string) error {
		runeRange, property, err := parseEntry(line)
		if err != nil {
			return err
		}
		return callback(runeRange, property)
	})
}

func LoadGeneralCategoryMap(filename string, dbInfoList *DataHeaders) (setMap UniSetMap[GeneralCategory], e error) {
	builderMap := map[GeneralCategory]*set.UniSetBuilder{}
	for cate := range EachGeneralCategory {
		builderMap[cate] = &set.UniSetBuilder{}
	}

	// load
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, err
	}
	err = loader.LoadProperties(func(runeRange set.RuneRange, property string) error {
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

func LoadEastAsianWidthMap(filename string, dbInfoList *DataHeaders) (setMap UniSetMap[EastAsianWidth], e error) {
	builderMap := map[EastAsianWidth]*set.UniSetBuilder{}
	for eaw := range EachEastAsianWidth {
		if eaw == EAW_N {
			continue // fill N later
		}
		builderMap[eaw] = &set.UniSetBuilder{}
	}

	// load
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, err
	}
	err = loader.LoadProperties(func(runeRange set.RuneRange, property string) error {
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

func LoadTargetAliasMap(filename string, dbInfoList *DataHeaders) (*AliasMapRecord, error) {
	aliasMapRecord := NewAliasMapRecord()
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, err
	}
	err = loader.Load(func(line string) error {
		return aliasMapRecord.Resolve(line)
	})
	if err != nil {
		return nil, err
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return aliasMapRecord, nil
}

func LoadScriptMap(filename string, aliasMap *AliasMap, dbInfoList *DataHeaders) (def *ScriptDef, setMap UniSetMap[Script], e error) {
	builderMap := map[Script]*set.UniSetBuilder{}
	nameToScript := map[string]Script{}

	// load
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, nil, err
	}
	err = loader.LoadProperties(func(runeRange set.RuneRange, property string) error {
		if _, ok := nameToScript[property]; !ok { // init
			script := Script(len(nameToScript))
			nameToScript[property] = script
			builderMap[script] = &set.UniSetBuilder{}
		}
		script := nameToScript[property]
		builderMap[script].AddRange(runeRange)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// fix-up
	longs := make([]string, len(nameToScript))
	for k, v := range nameToScript {
		longs[v] = k
	}
	scriptDef := NewScriptDef(longs, aliasMap)

	// build
	setMap = map[Script]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap[cate] = &tmp
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return scriptDef, setMap, nil
}

func LoadScriptXMap(filename string, def *ScriptDef, aliasMap *AliasMap, dbInfoList *DataHeaders) (setMap UniSetMap[Script], e error) {
	builderMap := map[Script]*set.UniSetBuilder{}

	// load
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, err
	}
	err = loader.LoadProperties(func(runeRange set.RuneRange, property string) error {
		ss := strings.Split(property, " ")
		for _, s := range ss {
			script, err := def.Parse(s, aliasMap)
			if err != nil {
				return err
			}
			if _, ok := builderMap[script]; !ok { // lazy init
				builderMap[script] = &set.UniSetBuilder{}
			}
			builderMap[script].AddRange(runeRange)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// build
	setMap = map[Script]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap[cate] = &tmp
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return setMap, nil
}

func LoadPropListMap(filename string, dbInfoList *DataHeaders) (def *PropertyDef[PropList], setMap UniSetMap[PropList], e error) {
	builderMap := map[PropList]*set.UniSetBuilder{}
	nameToProp := map[string]PropList{}

	// load
	loader, err := NewDataLoader(filename)
	if err != nil {
		return nil, nil, err
	}
	err = loader.LoadProperties(func(runeRange set.RuneRange, property string) error {
		if _, ok := nameToProp[property]; !ok { // init
			script := PropList(len(nameToProp))
			nameToProp[property] = script
			builderMap[script] = &set.UniSetBuilder{}
		}
		script := nameToProp[property]
		builderMap[script].AddRange(runeRange)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// fix-up
	longs := make([]string, len(nameToProp))
	for k, v := range nameToProp {
		longs[v] = k
	}
	propDef := NewPropertyDef(longs)

	// build
	setMap = map[PropList]*set.UniSet{}
	for cate, builder := range builderMap {
		tmp := builder.Build()
		setMap[cate] = &tmp
	}
	dbInfoList.List = append(dbInfoList.List, loader.header)
	return propDef, setMap, nil
}
