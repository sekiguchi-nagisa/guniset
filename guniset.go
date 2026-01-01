package main

import (
	"errors"
	"fmt"
	"io"
	"os"
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
	unicodeDir      string
	GeneralCategory io.ReadCloser // DerivedGeneralCategory.txt
	EastAsianWidth  io.ReadCloser // EastAsianWidth.txt
	Writer          io.Writer     // for generated Unicode set string
	SetOperation    string
}

func NewGUniSetFromDir(unicodeDir string, writer io.Writer, setOperation string) (*GUniSet, error) {
	generalCategory, err := os.Open(path.Join(unicodeDir, "DerivedGeneralCategory.txt"))
	if err != nil {
		return nil, err
	}
	eastAsianWidth, err := os.Open(path.Join(unicodeDir, "EastAsianWidth.txt"))
	if err != nil {
		return nil, err
	}
	return &GUniSet{
		unicodeDir:      unicodeDir,
		GeneralCategory: generalCategory,
		EastAsianWidth:  eastAsianWidth,
		Writer:          writer,
		SetOperation:    setOperation,
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

func (g *GUniSet) Run(filterOp SetFilterOp) error {
	ctx, err := op.NewEvalContext(g.GeneralCategory, g.EastAsianWidth)
	if err != nil {
		return err
	}
	node, err := op.NewParser().Run([]byte(g.SetOperation))
	if err != nil {
		return err
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
	return PrintUniSet(&uniSet, g.Writer)
}

func (g *GUniSet) Query() error {
	r, err := set.ParseRune(g.SetOperation)
	if err != nil {
		return err
	}
	ctx, err := op.NewEvalContext(g.GeneralCategory, g.EastAsianWidth)
	if err != nil {
		return err
	}
	return ctx.Query(r, g.Writer)
}

func (g *GUniSet) Info() error {
	ctx, err := op.NewEvalContext(g.GeneralCategory, g.EastAsianWidth)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(g.Writer, "GUNISET_DIR: %s\n", g.unicodeDir)
	if err != nil {
		return err
	}
	err = ctx.CateMap.PrintHeader(g.Writer)
	if err != nil {
		return err
	}
	err = ctx.EawMap.PrintHeader(g.Writer)
	if err != nil {
		return err
	}
	return nil
}

func (g *GUniSet) Close() error {
	err1 := g.GeneralCategory.Close()
	err2 := g.EastAsianWidth.Close()
	if err1 != nil || err2 != nil {
		return errors.Join(err1, err2)
	}
	return nil
}
