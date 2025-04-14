package main

import (
	"errors"
	"fmt"
	"github.com/sekiguchi-nagisa/guniset/op"
	"github.com/sekiguchi-nagisa/guniset/set"
	"io"
	"os"
	"path"
)

type GUniSet struct {
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
		GeneralCategory: generalCategory,
		EastAsianWidth:  eastAsianWidth,
		Writer:          writer,
		SetOperation:    setOperation,
	}, nil
}

func PrintUniSet(uniSet *set.UniSet, writer io.Writer) error {
	for interval := range uniSet.Interval {
		_, err := fmt.Fprintf(writer, "{ 0x%04X, 0x%04X },\n", interval.First, interval.Last)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GUniSet) Run() error {
	ctx, err := op.NewEvalContext(g.GeneralCategory, g.EastAsianWidth)
	if err != nil {
		return err
	}
	node, err := op.NewParser().Run([]byte(g.SetOperation))
	if err != nil {
		return err
	}
	uniSet := node.Eval(ctx)
	return PrintUniSet(&uniSet, g.Writer)
}

func (g *GUniSet) Close() error {
	err1 := g.GeneralCategory.Close()
	err2 := g.EastAsianWidth.Close()
	if err1 != nil || err2 != nil {
		return errors.Join(err1, err2)
	}
	return nil
}
