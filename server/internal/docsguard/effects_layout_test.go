package docsguard_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// splitEffectsDir is the tool that splits internal/cards/effects (#5906).
const splitEffectsDir = "../../../scripts/split-effects"

// TestEffectsTreeLayout is the fix-it check of #5906 (ADR 0224). The effects
// package is split into a core, card shards (cards/<shard>/) and test
// packages (tests/<id>/), and the frozen range table
// internal/cards/effects/shards.txt says which shard or test package each
// file name belongs to. A branch cut before the split adds its card and test
// files in the old place, internal/cards/effects/; they do not compile there,
// and this test says what to do about it: it runs
// `scripts/split-effects --check`, which lists every misplaced file and prints
// the one command that moves them and rewrites their names.
//
// Before the split (no core_cards.txt yet) it only checks that shards.txt
// parses: rows sorted, ids unique, each kind starting at 0.
func TestEffectsTreeLayout(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "--check", "-root", root)
	cmd.Dir = splitEffectsDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scripts/split-effects --check failed (%v):\n\n%s", err, out)
	}
}
