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

func (g *GUniSet) RunAndSampling(seed uint64, filterOp SetFilterOp, limit *int, ratio *float64) error {
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
		_, _ = fmt.Fprintf(g.Writer, "U+%04X\n", r)
	}
	return nil
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

func (g *GUniSet) Query() error {
	r, err := set.ParseRune(g.SetOperation)
	if err != nil {
		return err
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
	if op.IsGeneralCategoryPrefix(g.SetOperation) {
		for cat := range op.EachGeneralCategoryAll {
			_, _ = fmt.Fprintln(g.Writer, cat.Format(ctx.AliasMapRecord.Category()))
		}
		return nil
	}
	if op.IsEastAsianWidthPrefix(g.SetOperation) {
		for eaw := range op.EachEastAsianWidth {
			_, _ = fmt.Fprintln(g.Writer, eaw.Format(ctx.AliasMapRecord.Eaw()))
		}
		return nil
	}
	if op.IsScriptPrefix(g.SetOperation) || op.IsScriptExtensionPrefix(g.SetOperation) {
		for sc := range ctx.DefRecord.ScriptDef.EachScript {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.ScriptDef.Format(sc, ctx.AliasMapRecord.Script()))
		}
		return nil
	}
	if op.IsPropListPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.PropListDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.PropListDef.Format(prop))
		}
		return nil
	}
	if op.IsDerivedCorePropertyPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.DerivedCorePropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedCorePropDef.Format(prop))
		}
		return nil
	}
	if op.IsEmojiPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.EmojiDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.EmojiDef.Format(prop))
		}
		return nil
	}
	if op.IsDerivedBinaryPropertyPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.DerivedBinaryPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedBinaryPropDef.Format(prop))
		}
		return nil
	}
	if op.IsDerivedNormalizationPropPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.DerivedNormalizationPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.DerivedNormalizationPropDef.Format(prop))
		}
		return nil
	}
	if op.IsGraphemeBreakPropertyPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.GraphemeBreakPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.GraphemeBreakPropDef.Format(prop))
		}
		return nil
	}
	if op.IsWordBreakPropertyPrefix(g.SetOperation) {
		for prop := range ctx.DefRecord.WordBreakPropDef.EachProperty {
			_, _ = fmt.Fprintln(g.Writer, ctx.DefRecord.WordBreakPropDef.Format(prop))
		}
		return nil
	}
	if op.IsSentenceBreakPropertyPrefix(g.SetOperation) {
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
	return nil
}
