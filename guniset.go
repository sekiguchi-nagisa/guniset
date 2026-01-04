package main

import (
	"fmt"
	"io"
	"path"

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
	GeneralCategory      string    // DerivedGeneralCategory.txt
	EastAsianWidth       string    // EastAsianWidth.txt
	Scripts              string    // Scripts.txt
	ScriptExtensions     string    // ScriptExtensions.txt
	PropertyValueAliases string    // PropertyValueAliases.txt
	Writer               io.Writer // for generated Unicode set string
	SetOperation         string
}

func NewGUniSetFromDir(unicodeDir string, writer io.Writer, setOperation string) (*GUniSet, error) {
	return &GUniSet{
		GeneralCategory:      path.Join(unicodeDir, "DerivedGeneralCategory.txt"),
		EastAsianWidth:       path.Join(unicodeDir, "EastAsianWidth.txt"),
		Scripts:              path.Join(unicodeDir, "Scripts.txt"),
		ScriptExtensions:     path.Join(unicodeDir, "ScriptExtensions.txt"),
		PropertyValueAliases: path.Join(unicodeDir, "PropertyValueAliases.txt"),
		Writer:               writer,
		SetOperation:         setOperation,
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
	return op.NewEvalContext(g.GeneralCategory, g.EastAsianWidth, g.PropertyValueAliases, g.Scripts)
}

func (g *GUniSet) Run(filterOp SetFilterOp) (*set.UniSet, error) {
	ctx, err := g.prepare()
	if err != nil {
		return nil, err
	}
	node, err := op.NewParser(ctx.AliasMaps, ctx.ScriptDef).Run([]byte(g.SetOperation))
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

func (g *GUniSet) RunAndSampling(filterOp SetFilterOp, limit int) error {
	uniSet, err := g.Run(filterOp)
	if err != nil {
		return err
	}
	sampled := uniSet.Sample(limit)
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
	_, err = fmt.Fprintf(g.Writer, "GUNISET_DIR: %s\n", path.Dir(g.GeneralCategory))
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
			_, _ = fmt.Fprintln(g.Writer, cat.Format(ctx.AliasMaps[op.GeneralCategoryPrefix]))
		}
		return nil
	}
	if op.IsEastAsianWidthPrefix(g.SetOperation) {
		for eaw := range op.EachEastAsianWidth {
			_, _ = fmt.Fprintln(g.Writer, eaw.Format(ctx.AliasMaps[op.EastAsianWidthPrefix]))
		}
		return nil
	}
	if op.IsScriptPrefix(g.SetOperation) {
		for sc := range ctx.ScriptDef.EachScript {
			_, _ = fmt.Fprintln(g.Writer, ctx.ScriptDef.Format(sc, ctx.AliasMaps[op.ScriptPrefix]))
		}
		return nil
	}
	return fmt.Errorf("unknown property: %s", g.SetOperation)
}
