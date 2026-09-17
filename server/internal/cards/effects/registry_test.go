package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRegisterAndLookup covers the baseline registry contract:
// Register adds an entry; Lookup retrieves it; Has reports
// membership. The registry is process-lifetime, so each test uses
// a unique ScryfallID to avoid collisions with sibling tests.
func TestRegisterAndLookup(t *testing.T) {
	const id = "test-registry-baseline"
	Register(Spec{OracleID: id, Name: "Baseline"})
	spec, ok := Lookup(id)
	if !ok {
		t.Fatalf("Lookup(%q): ok=false, want true", id)
	}
	if spec.Name != "Baseline" {
		t.Errorf("Lookup(%q).Name: got %q, want %q", id, spec.Name, "Baseline")
	}
	if !Has(id) {
		t.Errorf("Has(%q): false, want true", id)
	}
}

// TestLookupMiss covers the non-catalog path: an unknown ID
// returns ok=false and Has returns false. The resolution path
// keys its auto-vs-manual decision off this signal.
func TestLookupMiss(t *testing.T) {
	const id = "test-registry-nonexistent-1234"
	if _, ok := Lookup(id); ok {
		t.Errorf("Lookup(%q): ok=true, want false (non-registered ID)", id)
	}
	if Has(id) {
		t.Errorf("Has(%q): true, want false", id)
	}
}

// TestRegisterDuplicatePanics pins the copy-paste catch: two Specs
// with the same ScryfallID panic at registration, surfacing the
// bug at server boot rather than silently letting one win.
func TestRegisterDuplicatePanics(t *testing.T) {
	const id = "test-registry-duplicate"
	Register(Spec{OracleID: id, Name: "First"})
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("duplicate Register did not panic")
		}
	}()
	Register(Spec{OracleID: id, Name: "Second"})
}

// TestRegisterEmptyIDPanics pins the sanity check: a Spec with no
// key would shadow Lookup("") and mask caller bugs.
func TestRegisterEmptyIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("empty-ID Register did not panic")
		}
	}()
	Register(Spec{Name: "MissingID"})
}

// TestRegisterTooManyReplacementsPanics pins #801's bound: a
// catalog ReplacementEffectID packs the source's battlefield index
// and its slot into one number, so a Spec declaring more slots than
// the stride would mint IDs belonging to the next permanent along.
// Unreachable from any printed card — which is why it is cheap to
// make impossible at boot.
func TestRegisterTooManyReplacementsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("over-budget Register did not panic")
		}
	}()
	Register(Spec{
		OracleID:     "test-registry-replacement-budget",
		Name:         "Slot Hog",
		Replacements: make([]game.ReplacementEffect, game.MaxCatalogReplacementSlots+1),
	})
}

// TestRegisterAtTheReplacementBudgetIsFine is the other side of the
// bound: the budget is inclusive, so a Spec exactly at it registers.
func TestRegisterAtTheReplacementBudgetIsFine(t *testing.T) {
	Register(Spec{
		OracleID:     "test-registry-replacement-budget-exact",
		Name:         "Slot Hog Jr",
		Replacements: make([]game.ReplacementEffect, game.MaxCatalogReplacementSlots),
	})
	if !Has("test-registry-replacement-budget-exact") {
		t.Errorf("a Spec at the slot budget did not register")
	}
}

// TestAllReturnsSnapshot proves All returns a fresh slice —
// mutating it does not leak back into the registry.
func TestAllReturnsSnapshot(t *testing.T) {
	const id = "test-registry-all-snapshot"
	Register(Spec{OracleID: id, Name: "SnapshotProbe"})
	snap := All()
	found := false
	for _, s := range snap {
		if s.OracleID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("All() did not include %q", id)
	}
	before := len(All())
	snap[0] = Spec{} // mutate caller-owned slice
	after := len(All())
	if before != after {
		t.Errorf("mutating All() leaked back: before=%d, after=%d", before, after)
	}
}
