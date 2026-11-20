// Command split-effects splits server/internal/cards/effects into a core,
// card shards and test packages (#5906, ADR 0225), and keeps the tree that
// way afterwards. Run it through scripts/split-effects.sh, from anywhere in
// the repository:
//
//	scripts/split-effects.sh                 split the tree (on a tree that is already split,
//	                                         it places any file --check reports, so it is idempotent)
//	scripts/split-effects.sh --place FILE…   move FILEs (card, helper or test files written for the
//	                                         one-package layout) where the layout puts them, rewriting
//	                                         the names they use; relative paths are read from the
//	                                         directory you run it in
//	scripts/split-effects.sh --check         list misplaced files and the one command that fixes them
//	                                         (what the docsguard test runs); exit 1 if there are any
//	scripts/split-effects.sh --plan          print what a split would do, write nothing
//
// It reads server/internal/cards/effects/shards.txt (the frozen range table),
// scripts/split-effects/overrides.txt (the reviewed names for collisions) and,
// after the split, scripts/split-effects/table.txt (every name the split
// changed, which --place uses). REGENERATE.md is the runbook.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootFlag := flag.String("root", "", "repository root (default: found from the working directory)")
	place := flag.Bool("place", false, "move the files named as arguments into place")
	checkMode := flag.Bool("check", false, "report misplaced files")
	plan := flag.Bool("plan", false, "print the split plan only")
	noBaseline := flag.Bool("no-baseline", false, "split: do not regenerate the clone baseline (go test -update)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: scripts/split-effects.sh [--plan | --check | --place FILE...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot()
		if err != nil {
			fatal(err)
		}
	}
	root, _ = filepath.Abs(root)
	var err error
	switch {
	case *checkMode:
		var ms []misplaced
		ms, err = check(root)
		if err == nil && len(ms) > 0 {
			fmt.Print(fixItMessage(ms))
			os.Exit(1)
		}
	case *place:
		if flag.NArg() == 0 {
			fatal(fmt.Errorf("--place needs the files to place"))
		}
		err = placeFiles(root, flag.Args())
	default:
		err = split(root, *plan, !*noBaseline)
	}
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "split-effects:", err)
	os.Exit(2)
}

// findRoot walks up from the working directory to the directory holding
// server/go.mod.
func findRoot() (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "server", "go.mod")); err == nil {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			return "", fmt.Errorf("no server/go.mod above the working directory; pass -root")
		}
		d = p
	}
}

func toolDir(root string) string { return filepath.Join(root, "scripts", "split-effects") }

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return strings.ReplaceAll(r, string(filepath.Separator), "/")
}
