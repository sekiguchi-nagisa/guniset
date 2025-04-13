package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

var gUniSetDir string

func fetchContent(url string, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	_, err = file.Write(body)
	return err
}

func TestMain(m *testing.M) {
	// get unicode data
	outputDir, err := os.MkdirTemp("", "guniset_test")
	if err != nil {
		log.Fatal(err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(outputDir)
	gUniSetDir = outputDir

	targets := []string{"UnicodeData.txt", "EastAsianWidth.txt"}
	rev := "16.0.0"
	for _, target := range targets {
		url := fmt.Sprintf("https://www.unicode.org/Public/%s/ucd/%s", rev, target)
		err = fetchContent(url, path.Join(outputDir, target))
		if err != nil {
			log.Fatal(err)
		}
	}

	exitCode := m.Run()
	if exitCode == 0 {
		_ = os.RemoveAll(outputDir)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "@@ failed workdir: %s\n", outputDir)
	}
	os.Exit(exitCode)
}

func TestRun(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	targetDir := path.Join(wd, "test", "unicode16")
	cases, err := filepath.Glob(path.Join(targetDir, "*_test.txt"))
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			testData, err := os.ReadFile(c)
			if err != nil {
				t.Fatal(err)
			}
			expectFile := strings.TrimSuffix(c, "_test.txt") + "_golden.txt"
			expectData, err := os.ReadFile(expectFile)
			if err != nil {
				t.Fatal(err)
			}

			actualWriter := strings.Builder{}
			g, err := NewGUniSetFromDir(gUniSetDir, &actualWriter, string(testData))
			if err != nil {
				t.Fatal(err)
			}
			defer func(g *GUniSet) {
				_ = g.Close()
			}(g)
			err = g.Run()
			if err != nil {
				t.Fatal(err)
			}

			// compare result
			assert.Equal(t, strings.TrimSpace(string(expectData)),
				strings.TrimSpace(actualWriter.String()), "See "+expectFile)
		})
	}
}
