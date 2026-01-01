package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
)

type CLIGen struct {
	Set    string `arg:"" required:"" help:"Specify set operation"`
	Filter string `optional:"" help:"Filter output (all: include all, bmp: only bmp, non-bmp: exclude bmp)" enum:"all,,bmp,non-bmp" default:"all"`
}

type CLIQuery struct {
	CodePoint string `arg:"" required:"" help:"Specify code point to query"`
}

type CLIInfo struct {
}

type CLISample struct {
	Set    string `arg:"" required:"" help:"Specify set operation"`
	Filter string `optional:"" help:"Filter output (all: include all, bmp: only bmp, non-bmp: exclude bmp)" enum:"all,,bmp,non-bmp" default:"all"`
	Limit  int    `optional:"" help:"Limit sampling count" default:"5"`
}

var CLI struct {
	Version  kong.VersionFlag `short:"v" help:"Show version information"`
	Generate CLIGen           `cmd:"" help:"Generate Unicode set"`
	Query    CLIQuery         `cmd:"" help:"Query code point property"`
	Info     CLIInfo          `cmd:"" help:"Show information about Unicode database"`
	Sample   CLISample        `cmd:"" help:"Sample Unicode code points"`
}

var version = "" // for version embedding (specified like "-X main.version=v0.1.0")

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		rev := "unknown"
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				rev = setting.Value
				break
			}
		}
		var v = info.Main.Version
		if version != "" { // set by "-X main.version=v0.1.0"
			v = version
		}
		return fmt.Sprintf("%s (%s)", v, rev)
	} else {
		return "(unknown)"
	}
}

func resolveGunisetDir() (string, error) {
	gunisetDir := os.Getenv("GUNISET_DIR")
	if gunisetDir == "" {
		gunisetDir = "."
	}
	_, err := os.Stat(gunisetDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve GUNISET_DIR: %v", err)
	}
	return gunisetDir, nil
}

func (c *CLIGen) Run() error {
	gunisetDir, err := resolveGunisetDir()
	if err != nil {
		return err
	}
	g, err := NewGUniSetFromDir(gunisetDir, os.Stdout, c.Set)
	if err != nil {
		return err
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)
	printOp, ok := StrToSetPrintOps[c.Filter]
	if !ok {
		return fmt.Errorf("unknown filter %q\n", c.Filter)
	}
	return g.RunAndPrint(printOp)
}

func (c *CLIQuery) Run() error {
	gunisetDir, err := resolveGunisetDir()
	if err != nil {
		return err
	}
	g, err := NewGUniSetFromDir(gunisetDir, os.Stdout, c.CodePoint)
	if err != nil {
		return err
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)
	return g.Query()
}

func (c *CLIInfo) Run() error {
	gunisetDir, err := resolveGunisetDir()
	if err != nil {
		return err
	}
	g, err := NewGUniSetFromDir(gunisetDir, os.Stdout, "")
	if err != nil {
		return err
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)
	return g.Info()
}

func (c *CLISample) Run() error {
	gunisetDir, err := resolveGunisetDir()
	if err != nil {
		return err
	}
	g, err := NewGUniSetFromDir(gunisetDir, os.Stdout, c.Set)
	if err != nil {
		return err
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)
	printOp, ok := StrToSetPrintOps[c.Filter]
	if !ok {
		return fmt.Errorf("unknown filter %q\n", c.Filter)
	}
	return g.RunAndSampling(printOp, c.Limit)
}

func main() {
	ctx := kong.Parse(&CLI, kong.UsageOnError(), kong.Vars{"version": getVersion()})
	err := ctx.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
