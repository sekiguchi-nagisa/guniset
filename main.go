package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"io"
	"os"
	"runtime/debug"
)

var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Show version information"`
	Output  string           `short:"o" help:"Set output file (default stdout)"`
	Set     string           `arg:"" required:"" help:"Specify set operation"`
	Filter  string           `optional:"" help:"Filter output (all: include all, bmp: only bmp, non-bmp: exclude bmp)" enum:"all,,bmp,non-bmp" default:"all"`
	Query   bool             `short:"q" optional:"" help:"Query code point property"`
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

func main() {
	kong.Parse(&CLI, kong.UsageOnError(), kong.Vars{"version": getVersion()})
	gunisetDir := os.Getenv("GUNISET_DIR")
	if gunisetDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot get current directory: %v\n", err)
			os.Exit(1)
		}
		gunisetDir = dir
	}
	var writer io.Writer = os.Stdout
	if CLI.Output != "" {
		w, err := os.Create(CLI.Output)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot open output file: %v\n", err)
			os.Exit(1)
		}
		writer = w
	}
	g, err := NewGUniSetFromDir(gunisetDir, writer, CLI.Set)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)

	if CLI.Query {
		err = g.Query()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	printOp, ok := StrToSetPrintOps[CLI.Filter]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "unknown filter %q\n", CLI.Filter)
		os.Exit(1)
	}
	err = g.Run(printOp)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
