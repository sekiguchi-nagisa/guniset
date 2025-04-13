package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"io"
	"log"
	"os"
	"runtime/debug"
)

var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Show version information"`
	Output  string           `short:"o" help:"Set output file (default stdout)"`
	Set     string           `arg:"" required:"" help:"Specify set operation"`
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
			log.Fatalf("cannot get current directory: %v", err)
		}
		gunisetDir = dir
	}
	var writer io.Writer = os.Stdout
	if CLI.Output != "" {
		w, err := os.Create(CLI.Output)
		if err != nil {
			log.Fatalf("cannot open output file: %v", err)
		}
		writer = w
	}
	g, err := NewGUniSetFromDir(gunisetDir, writer, CLI.Set)
	if err != nil {
		log.Fatal(err)
	}
	defer func(g *GUniSet) {
		_ = g.Close()
	}(g)
	err = g.Run()
	if err != nil {
		log.Fatal(err)
	}
}
