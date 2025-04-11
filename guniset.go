package main

import (
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

func (g *GUniSet) Run() error {

	return nil
}
