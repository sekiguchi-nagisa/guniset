package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/sekiguchi-nagisa/guniset/op"
	"github.com/sekiguchi-nagisa/guniset/set"
)

type SetFilterOp int8

const (
	SetPrintAll SetFilterOp = iota
	SetPrintBMP
	SetPrintNonBMP
)

var StrToSetPrintOps = map[string]SetFilterOp{
	"all":     SetPrintAll,
	"bmp":     SetPrintBMP,
	"non-bmp": SetPrintNonBMP,
}

type GUniSet struct {
	UnicodeData  *op.UnicodeData
	Writer       io.Writer // for generated Unicode set string
	SetOperation string
}

func NewGUniSetFromDir(unicodeDir string, writer io.Writer, setOperation string) (*GUniSet, error) {
	return &GUniSet{
		UnicodeData:  op.NewUnicodeData(unicodeDir),
		Writer:       writer,
		SetOperation: setOperation,
	}, nil
}

func PrintUniSet(uniSet *set.UniSet, writer io.Writer) error {
	for runeRange := range uniSet.Range {
		_, err := fmt.Fprintf(writer, "{ 0x%04X, 0x%04X },\n", runeRange.First, runeRange.Last)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GUniSet) prepare() (*op.EvalContext, error) {
	return op.NewEvalContext(g.UnicodeData)
}

func (g *GUniSet) Run(filterOp SetFilterOp) (*set.UniSet, error) {
	ctx, err := g.prepare()
	if err != nil {
		return nil, err
	}
	node, err := op.NewParser(ctx.AliasMapRecord, &ctx.DefRecord).Run([]byte(g.SetOperation))
	if err != nil {
		return nil, err
	}
	uniSet := node.Eval(ctx)
	switch filterOp {
	case SetPrintAll: // do nothing
	case SetPrintBMP:
		uniSet.Filter(func(r rune) bool { // only allow bmp rune (remove non-bmp)
			return set.IsBmpRune(r)
		})
	case SetPrintNonBMP:
		uniSet.Filter(func(r rune) bool { // only allow non-bmp rune (remove bmp)
			return set.IsSupplementaryRune(r)
		})
	}
	return &uniSet, nil
}

func (g *GUniSet) RunAndPrint(filterOp SetFilterOp) error {
	uniSet, err := g.Run(filterOp)
	if err != nil {
		return err
	}
	return PrintUniSet(uniSet, g.Writer)
}

type PrintFormat int8

const (
	CodePointFormat PrintFormat = iota
	StringFormat
	Utf8EscapeFormat
)

var strToPrintFormat = map[string]PrintFormat{
	"codepoint":  CodePointFormat,
	"string":     StringFormat,
	"utf8escape": Utf8EscapeFormat,
}

func stringToUtf8Escape(s string) string {
	sb := strings.Builder{}
	for i := 0; i < len(s); i++ {
		b := s[i]
		sb.WriteString(fmt.Sprintf("\\x%02X", b))
	}
	return sb.String()
}

func (g *GUniSet) RunAndSampling(seed uint64, filterOp SetFilterOp, format PrintFormat, limit *int, ratio *float64) error {
	uniSet, err := g.Run(filterOp)
	if err != nil {
		return err
	}
	actualLimit := 5
	if limit != nil {
		actualLimit = *limit
	} else if ratio != nil {
		actualLimit = int(*ratio * float64(uniSet.Len()))
	}

	rnd := rand.New(rand.NewPCG(seed, 42))
	sampled := uniSet.Sample(rnd, actualLimit)
	for r := range sampled.Iter {
		switch format {
		case CodePointFormat:
			_, _ = fmt.Fprintf(g.Writer, "U+%04X\n", r)
		case StringFormat:
			_, _ = fmt.Fprintf(g.Writer, "%s\n", string(r))
		case Utf8EscapeFormat:
			_, _ = fmt.Fprintf(g.Writer, "%s\n", stringToUtf8Escape(string(r)))
		}
	}
	return nil
}

func (g *GUniSet) RunStrings(format PrintFormat) error {
	ctx, err := op.NewEvalContext(g.UnicodeData)
	if err != nil {
		return err
	}
	if g.SetOperation == "" {
		list := op.Properties(ctx.StringPropertyMap)
		_, _ = fmt.Fprintf(g.Writer, "must be: %s\n", strings.Join(list, ", "))
		return nil
	}
	if values := op.LookupStringPropertyValues(ctx.StringPropertyMap, g.SetOperation); len(values) > 0 {
		for _, v := range values {
			switch format {
			case CodePointFormat:
				for i, r := range v.Runes() {
					if i > 0 {
						_, _ = fmt.Fprintf(g.Writer, " ")
					}
					_, _ = fmt.Fprintf(g.Writer, "U+%04X", r)
				}
				_, _ = fmt.Fprintf(g.Writer, "\n")
			case StringFormat:
				_, _ = fmt.Fprintf(g.Writer, "%s\n", v.String())
			case Utf8EscapeFormat:
				_, _ = fmt.Fprintf(g.Writer, "%s\n", stringToUtf8Escape(v.String()))
			}
		}
		return nil
	}
	list := op.Properties(ctx.StringPropertyMap)
	return fmt.Errorf("unsupported string property: %s\nmust be: %s",
		g.SetOperation, strings.Join(list, ", "))
}

func formatScriptX(def *op.ScriptDef, scx []op.Script) string {
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

func formatEmoji(def *op.PropertyDef[op.Emoji], emoji []op.Emoji) string {
	builder := strings.Builder{}
	builder.WriteString("[")
	for i, s := range emoji {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(def.Format(s))
	}
	builder.WriteString("]")
	return builder.String()
}

func (g *GUniSet) Query(asString bool) error {
	var r rune
	if asString {
		rr := []rune(g.SetOperation)
		if len(rr) != 1 {
			return errors.New("invalid string. must be exactly one rune")
		}
		r = rr[0]
	} else {
		r1, err := set.ParseRune(g.SetOperation)
		if err != nil {
			return err
		}
		r = r1
	}
	ctx, err := g.prepare()
	if err != nil {
		return err
	}
	cat := op.CAT_Cn
	eaw := op.EAW_N
	sc := ctx.DefRecord.ScriptDef.Unknown()
	var emoji []op.Emoji
	gbp := ""
	wbp := ""
	sbp := ""
	var scx []op.Script
	for cc, uniSet := range ctx.CateMap {
		if uniSet.Find(r) {
			cat = cc
			break
		}
	}
	for e, uniSet := range ctx.EawMap {
		if uniSet.Find(r) {
			eaw = e
			break
		}
	}
	for s, uniSet := range ctx.ScriptMap {
		if uniSet.Find(r) {
			sc = s
			break
		}
	}
	for s := range ctx.DefRecord.ScriptDef.EachScript {
		if m, ok := ctx.ScriptXMap[s]; ok && m.Find(r) {
			scx = append(scx, s) // may have multiple property
		}
	}
	for s := range ctx.DefRecord.EmojiDef.EachProperty {
		if m, ok := ctx.EmojiMap[s]; ok && m.Find(r) {
			emoji = append(emoji, s) // may have multiple property
		}
	}
	for s, uniSet := range ctx.GraphemeBreakPropMap {
		if uniSet.Find(r) {
			gbp = ctx.DefRecord.GraphemeBreakPropDef.Format(s)
			break
		}
	}
	for s, uniSet := range ctx.WordBreakPropMap {
		if uniSet.Find(r) {
			wbp = ctx.DefRecord.WordBreakPropDef.Format(s)
			break
		}
	}
	for s, uniSet := range ctx.SentenceBreakPropMap {
		if uniSet.Find(r) {
			sbp = ctx.DefRecord.SentenceBreakPropDef.Format(s)
		}
	}
	_, err = fmt.Fprintf(g.Writer, "CodePoint: U+%04X\n"+
		"GeneralCategory: %s\n"+
		"EastAsianWidth: %s\n"+
		"Script: %s\n"+
		"ScriptExtension: %s\n"+
		"Emoji: %s\n"+
		"GraphemeBreak: %s\n"+
		"WordBreak: %s\n"+
		"SentenceBreak: %s\n", r,
		cat.Format(ctx.AliasMapRecord.Category()),
		eaw.Format(ctx.AliasMapRecord.Eaw()),
		ctx.DefRecord.ScriptDef.Format(sc, ctx.AliasMapRecord.Script()),
		formatScriptX(ctx.DefRecord.ScriptDef, scx),
		formatEmoji(ctx.DefRecord.EmojiDef, emoji),
		gbp, wbp, sbp)
	return err
}

func (g *GUniSet) Info() error {
	ctx, err := g.prepare()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(g.Writer, "GUNISET_DIR: %s\n", path.Dir(g.UnicodeData.GeneralCategory))
	if err != nil {
		return err
	}
	return ctx.Headers.Print(g.Writer)
}

func (g *GUniSet) EnumerateProperty() error {
	ctx, err := g.prepare()
	if err != nil {
		return err
	}
	switch {
	case op.IsGeneralCategoryPrefix(g.SetOperation):
		for cat := range op.EachGeneralCategoryAll {
			_, _ = fmt.Fprintln(g.Writer, cat.Format(ctx.AliasMapRecord.Category()))
		}
		return nil
	case op.IsEastAsianWidthPrefix(g.SetOperation):
		for eaw := range op.EachEastAsianWidth {
			_, _ = fmt.Fprintln(g.Writer, eaw.Format(ctx.AliasMapRecord.Eaw()))
		}
		return nil
	case op.IsScriptPrefix(g.SetOperation) || op.IsScriptExtensionPrefix(g.SetOperation):
		for sc := range ctx.DefRecord.ScriptDef.EachScript {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.ScriptDef.Format(sc, ctx.AliasMapRecord.Script()))
		}
		return nil
	case op.IsPropListPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.PropListDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.PropListDef.Format(prop))
		}
		return nil
	case op.IsDerivedCorePropertyPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.DerivedCorePropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedCorePropDef.Format(prop))
		}
		return nil
	case op.IsEmojiPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.EmojiDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.EmojiDef.Format(prop))
		}
		return nil
	case op.IsDerivedBinaryPropertyPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.DerivedBinaryPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedBinaryPropDef.Format(prop))
		}
		return nil
	case op.IsDerivedNormalizationPropPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.DerivedNormalizationPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedNormalizationPropDef.Format(prop))
		}
		return nil
	case op.IsGraphemeBreakPropertyPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.GraphemeBreakPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.GraphemeBreakPropDef.Format(prop))
		}
		return nil
	case op.IsWordBreakPropertyPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.WordBreakPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.WordBreakPropDef.Format(prop))
		}
		return nil
	case op.IsSentenceBreakPropertyPrefix(g.SetOperation):
		for prop := range ctx.DefRecord.SentenceBreakPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.SentenceBreakPropDef.Format(prop))
		}
		return nil
	}
	return errors.New(op.UnknowPropertyPrefixError(g.SetOperation))
}

func fetchContent(url string, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot fetch %s: %v", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot fetch %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body %s: %v", url, err)
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = file.Write(body)
	return err
}

var revPattern = regexp.MustCompile(`^[1-9][0-9]+\.[0-9]+\.[0-9]+$`)

type Revision struct {
	major, minor, patch int
}

func NewRevision(rev string) (*Revision, error) {
	rev1s := strings.Split(rev, ".")
	if len(rev1s) != 3 {
		return nil, fmt.Errorf("invalid revision %q", rev)
	}
	major, err := strconv.Atoi(rev1s[0])
	if err != nil {
		return nil, fmt.Errorf("invalid revision %q", rev)
	}
	minor, err := strconv.Atoi(rev1s[1])
	if err != nil {
		return nil, fmt.Errorf("invalid revision %q", rev)
	}
	patch, err := strconv.Atoi(rev1s[2])
	if err != nil {
		return nil, fmt.Errorf("invalid revision %q", rev)
	}
	return &Revision{major, minor, patch}, nil
}

func (rev *Revision) Compare(rev2 *Revision) int {
	if rev.major != rev2.major {
		return rev.major - rev2.major
	}
	if rev.minor != rev2.minor {
		return rev.minor - rev2.minor
	}
	return rev.patch - rev2.patch
}

func compareRevision(rev1s string, rev2s string) int {
	rev1, err := NewRevision(rev1s)
	if err != nil {
		return -1
	}
	rev2, err := NewRevision(rev2s)
	if err != nil {
		return 1
	}
	return rev1.Compare(rev2)
}

func fetchUnicodeData(rev string, output string) error {
	if !revPattern.MatchString(rev) && rev != "latest" {
		return fmt.Errorf("invalid revision %q", rev)
	}

	targets := []string{
		"extracted/DerivedGeneralCategory.txt", "EastAsianWidth.txt", "PropertyValueAliases.txt",
		"Scripts.txt", "ScriptExtensions.txt", "PropList.txt", "DerivedCoreProperties.txt",
		"emoji/emoji-data.txt", "extracted/DerivedBinaryProperties.txt", "DerivedNormalizationProps.txt",
		"auxiliary/GraphemeBreakProperty.txt", "auxiliary/WordBreakProperty.txt", "auxiliary/SentenceBreakProperty.txt",
		"CaseFolding.txt",
	}
	if rev == "latest" {
		rev = "UCD/latest"
	}
	for _, target := range targets {
		url := fmt.Sprintf("https://www.unicode.org/Public/%s/ucd/%s", rev, target)
		log.Printf("@@ try downloading %s to %s", url, output)
		err := fetchContent(url, path.Join(output, path.Base(target)))
		if err != nil {
			return err
		}
	}

	// for emoji sequence
	targets = []string{
		"emoji-sequences.txt",
		"emoji-zwj-sequences.txt",
	}
	for _, target := range targets {
		var url string
		if rev == "UCD/latest" || compareRevision(rev, "17.0.0") >= 0 {
			url = fmt.Sprintf("https://www.unicode.org/Public/%s/emoji/%s", rev, target)
		} else {
			revs := strings.Split(rev, ".")
			url = fmt.Sprintf("https://www.unicode.org/Public/emoji/%s.%s/%s", revs[0], revs[1], target)
		}
		log.Printf("@@ try downloading %s to %s", url, output)
		err := fetchContent(url, path.Join(output, path.Base(target)))
		if err != nil {
			return err
		}
	}
	return nil
}
