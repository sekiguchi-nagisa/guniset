package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"regexp"

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

func (g *GUniSet) RunAndSampling(seed uint64, filterOp SetFilterOp, limit int) error {
	uniSet, err := g.Run(filterOp)
	if err != nil {
		return err
	}
	rnd := rand.New(rand.NewPCG(seed, 42))
	sampled := uniSet.Sample(rnd, limit)
	for r := range sampled.Iter {
		_, _ = fmt.Fprintf(g.Writer, "U+%04X\n", r)
	}
	return nil
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
	return ctx.Query(r, g.Writer)
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
	return fmt.Errorf("unknown property: %s", g.SetOperation)
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
		"emoji/emoji-data.txt",
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
