package main

// fixups.go — the hand fixes the rules cannot derive, as tool steps.
//
// Each is a text replacement in one moved test file (found by its base name
// anywhere in the effects tree), applied after the split. A fixup whose old
// text is not there fails the run, so a newer develop that changed the line
// is noticed instead of skipped: re-read the test, update the entry here.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fixup struct {
	file, old, new, why string
}

var fixups = []fixup{
	{
		file: "footprint_conformance_test.go",
		old:  `if got := fmt.Sprintf("%T", c.step); got != "effects."+name {`,
		new:  `if got := fmt.Sprintf("%T", c.step); got[strings.LastIndex(got, ".")+1:] != name {`,
		why:  "a footprint step type may now live in a card shard (cardsa.FoodOrTreasureStep), so match its name, not the package",
	},
}

func applyFixups(effDir string) error {
	paths := map[string]string{}
	err := filepath.WalkDir(effDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".go") {
			paths[d.Name()] = p
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, fx := range fixups {
		p, ok := paths[fx.file]
		if !ok {
			return fmt.Errorf("fixup for %s: no such file in the tree (%s)", fx.file, fx.why)
		}
		if err := replaceOnce(p, fx.old, fx.new); err != nil {
			return fmt.Errorf("fixup (%s): %v", fx.why, err)
		}
	}
	return nil
}
