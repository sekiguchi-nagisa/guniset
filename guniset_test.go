package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var gUniSetDir string

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

	err = fetchUnicodeData("16.0.0", outputDir)
	if err != nil {
		log.Fatal(err)
	}
	exitCode := m.Run()
	if exitCode == 0 {
		_ = os.RemoveAll(outputDir)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "@@ failed workdir: %s\n", outputDir)
	}
	os.Exit(exitCode)
}

func runGoldenTest(t *testing.T, baseName string, filterOp SetFilterOp) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	targetDir := path.Join(wd, "test", "generate", baseName)
	cases, err := filepath.Glob(path.Join(targetDir, "*.test"))
	if err != nil {
		t.Fatal(err)
	}
	if cases == nil || len(cases) == 0 {
		t.Fatalf("no test cases found in %s", targetDir)
	}

	t.Parallel()
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testData, err := os.ReadFile(c)
			if err != nil {
				t.Fatal(err)
			}
			expectFile := strings.TrimSuffix(c, ".test") + ".golden"
			expectData, err := os.ReadFile(expectFile)
			if err != nil {
				t.Fatal(err)
			}

			actualWriter := strings.Builder{}
			g, err := NewGUniSetFromDir(gUniSetDir, &actualWriter, string(testData))
			if err != nil {
				t.Fatal(err)
			}
			err = g.RunAndPrint(filterOp)
			if err != nil {
				t.Fatal(err)
			}

			// compare result
			assert.Equal(t, strings.TrimSpace(string(expectData)),
				strings.TrimSpace(actualWriter.String()), "See "+expectFile)
		})
	}
}

func TestPrintAll(t *testing.T) {
	runGoldenTest(t, "unicode16", SetPrintAll)
}

func TestPrintLong(t *testing.T) {
	runGoldenTest(t, "unicode16_long", SetPrintAll)
}

func TestPrintBMP(t *testing.T) {
	runGoldenTest(t, "unicode16_bmp", SetPrintBMP)
}

func TestPrintNonBMP(t *testing.T) {
	runGoldenTest(t, "unicode16_nonbmp", SetPrintNonBMP)
}

func TestPrintScript(t *testing.T) {
	runGoldenTest(t, "unicode16_script", SetPrintAll)
}

func TestPrintPropList(t *testing.T) {
	runGoldenTest(t, "unicode16_proplist", SetPrintAll)
}

func TestPrintDerivedCoreProperty(t *testing.T) {
	runGoldenTest(t, "unicode16_dcp", SetPrintAll)
}

func TestPrintEmoji(t *testing.T) {
	runGoldenTest(t, "unicode16_emoji", SetPrintAll)
}

func TestPrintDerivedBinaryProperties(t *testing.T) {
	runGoldenTest(t, "unicode16_dbp", SetPrintAll)
}

func TestPrintDerivedNormalizationProperties(t *testing.T) {
	runGoldenTest(t, "unicode16_dnp", SetPrintAll)
}
