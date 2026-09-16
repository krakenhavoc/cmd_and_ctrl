package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// carddef_test.go — #622. The engine reads one precomputed CardDef per
// card; the per-slot hooks derive from it and allocate nothing.

const (
	serraAngelOracleForDef    = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
	lightningBoltOracleForDef = "d4af6b58-f0b7-4c8f-8f4b-6f2a5a43e2c1"
)

func TestCardDefIsBuiltOnceWithTheKeywordStatic(t *testing.T) {
	d := game.CatalogLookup(serraAngelOracleForDef)
	if d == nil {
		t.Fatal("Serra Angel has no CardDef")
	}
	if len(d.Static) != 1 || d.Static[0].Layer != game.Layer6Ability {
		t.Fatalf("Serra Angel should carry exactly one synthesised Layer 6 keyword static, got %d", len(d.Static))
	}
	if len(d.PrintedKeywords) != 2 {
		t.Errorf("printed keywords %v, want flying + vigilance", d.PrintedKeywords)
	}
	// Two reads hand out the SAME slice: nothing is rebuilt per call.
	a, b := game.CatalogStaticAbilities(serraAngelOracleForDef), game.CatalogStaticAbilities(serraAngelOracleForDef)
	if &a[0] != &b[0] {
		t.Error("CatalogStaticAbilities rebuilt the keyword static on the second call")
	}
	if allocs := testing.AllocsPerRun(50, func() { _ = game.CatalogStaticAbilities(serraAngelOracleForDef) }); allocs != 0 {
		t.Errorf("CatalogStaticAbilities allocates %.0f per call, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(50, func() { _ = game.CatalogTriggers(serraAngelOracleForDef) }); allocs != 0 {
		t.Errorf("CatalogTriggers allocates %.0f per call, want 0", allocs)
	}
	if game.CatalogLookup("no-such-card") != nil || game.IsAutoCard("no-such-card") {
		t.Error("an unknown key must read as no catalog entry")
	}
}

func TestCardDefCarriesTheResolveWrapperAndTargetMode(t *testing.T) {
	spec, ok := Lookup(lightningBoltOracleForDef)
	if !ok {
		// The bolt's oracle ID moved; find it by name so the test does not rot.
		for _, s := range All() {
			if s.Name == "Lightning Bolt" {
				spec = s
				ok = true
			}
		}
	}
	if !ok {
		t.Skip("no Lightning Bolt in the catalog")
	}
	d := game.CatalogLookup(spec.OracleID)
	if d == nil || d.Resolve == nil {
		t.Fatal("a spell with OnResolve must have a Resolve wrapper")
	}
	if d.Targets == nil || d.TargetMode != d.Targets.Mode || game.TargetModeFor(spec.OracleID) != d.Targets.Mode {
		t.Errorf("TargetMode %q should be derived from the structured clause %q", d.TargetMode, d.Targets.Mode)
	}
	// Every registered card has a def, and the two registries agree.
	n := 0
	for _, s := range All() {
		if game.CatalogLookup(s.OracleID) == nil {
			t.Errorf("%s registered without a CardDef", s.Name)
		}
		n++
	}
	if n == 0 {
		t.Fatal("empty catalog")
	}
}
