package heuristic_test

import (
	"go/build"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// imports_test.go is the type gate from ADR 0033 §3, and it is the
// deliverable this package exists to carry as much as the policy is:
// the package doc on aiseat/policy.go promises that "real policies
// land in subpackages under aiseat/ that are forbidden from importing
// internal/game; the import test ships with the first of them".
// This is that test, and it protects every policy written after this
// one without those authors having to remember the rule.
//
// The guarantee is worth restating, because it is easy to read as
// bureaucracy. A Policy receives a protocol.GameView that has already
// been through FilterViewFor — opponents' hands are counts, their
// libraries are counts, face-down cards are backs. A policy holding a
// *game.Game would be reading the authoritative state instead, where
// every card in every hand is right there in a slice. No amount of
// care makes that safe; not having the handle does.
//
// The ban is on the DIRECT import only. protocol and legal both
// import internal/game — they are projections OF it — so a transitive
// ban is not expressible and would not mean anything if it were. What
// matters is that no policy names the package itself.

const bannedImport = "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exemptDirs are subpackages of aiseat/ that are not policies and
// legitimately touch the engine's types. Keep this list SHORT and
// justify every entry; a policy must never appear on it.
//
//   - decks: curated bot decks (S31 sub-PR 5) are []game.Card
//     fixtures handed to Game.AddPlayer at seat time. They are data,
//     they never see an Input, and they make no decisions.
var exemptDirs = map[string]bool{
	"decks": true,
}

func TestPolicyPackagesDoNotImportGame(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve aiseat root: %v", err)
	}

	checked := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata") {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			// aiseat itself is the runner: it drives the engine and
			// imports it on purpose. The ban is on what it calls.
			return nil
		}
		if exemptDirs[strings.Split(rel, string(filepath.Separator))[0]] {
			return filepath.SkipDir
		}
		pkg, err := build.ImportDir(path, 0)
		if err != nil {
			if _, noGo := err.(*build.NoGoError); noGo {
				return nil
			}
			return err
		}
		checked++
		for label, imports := range map[string][]string{
			"":      pkg.Imports,
			"test ": pkg.TestImports,
			"xtest": pkg.XTestImports,
		} {
			for _, imp := range imports {
				if imp == bannedImport {
					t.Errorf("aiseat/%s %simports %s — a policy package may never hold a *game.Game (ADR 0033 §3). "+
						"Everything it needs is on protocol.GameView and []legal.Move.", rel, label, bannedImport)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if checked == 0 {
		t.Fatal("no policy packages found under aiseat/ — the walk is broken, not the codebase")
	}
}

// TestImportBanIsLoadBearing is the control: if the ban ever stopped
// matching anything (a renamed module path, say), the test above
// would pass vacuously and nobody would notice for a year. The runner
// DOES import internal/game, so the constant must find it there.
func TestImportBanIsLoadBearing(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve aiseat root: %v", err)
	}
	pkg, err := build.ImportDir(root, 0)
	if err != nil {
		t.Fatalf("import aiseat: %v", err)
	}
	for _, imp := range pkg.Imports {
		if imp == bannedImport {
			return
		}
	}
	t.Fatalf("aiseat no longer imports %s — the import ban is matching a path that does not exist", bannedImport)
}
