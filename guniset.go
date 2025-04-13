package main

import (
	"fmt"
	"github.com/sekiguchi-nagisa/guniset/op"
	"github.com/sekiguchi-nagisa/guniset/set"
	"io"
	"os"
	"path"
)

type GUniSet struct {
	UnicodeData    io.Reader // UnicodeData.txt
	EastAsianWidth io.Reader // EastAsianWidth.txt
	Writer         io.Writer // for generated Unicode set string
	SetOperation   string
}

func NewGUniSetFromDir(unicodeDir string, writer io.Writer, setOperation string) (*GUniSet, error) {
	unicodeData, err := os.Open(path.Join(unicodeDir, "UnicodeData.txt"))
	if err != nil {
		return nil, err
	}
	eastAsianWidth, err := os.Open(path.Join(unicodeDir, "EastAsianWidth.txt"))
	if err != nil {
		return nil, err
	}
	return &GUniSet{
		UnicodeData:    unicodeData,
		EastAsianWidth: eastAsianWidth,
		Writer:         writer,
		SetOperation:   setOperation,
	}, nil
}

func PrintUniSet(uniSet *set.UniSet, writer io.Writer) error {
	count := 0
	for interval := range uniSet.Interval {
		if count > 0 {
			_, err := fmt.Fprint(writer, ",\n")
			if err != nil {
				return err
			}
		}
		count++
		_, err := fmt.Fprintf(writer, "{ 0x%04X, 0x%04X }", interval.First, interval.Last)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}

func (g *GUniSet) Run() error {
	ctx, err := op.NewEvalContext(g.UnicodeData, g.EastAsianWidth)
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
