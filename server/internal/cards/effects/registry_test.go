package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registerForTest registers spec and schedules its removal from the
// registry when the test completes. Register's map write is the only
// state a Spec leaves behind that a later pass of the same test can
// collide with (#1070): init()-time registrations run once no matter
// how many times `go test -count=N` re-executes the binary, but a
// Register call inside a test function body runs on every pass, and
// the package-level registry does not reset between them. A test that
// registers a fixed OracleID through this helper instead of calling
// Register directly cleans up after itself, so `-count=N` sees the
// same empty starting registry on every pass. See registry.go's
// package doc, which already names this helper.
func registerForTest(t *testing.T, spec Spec) {
	t.Helper()
	Register(spec)
	t.Cleanup(func() { delete(registry, spec.OracleID) })
}

// TestRegisterAndLookup covers the baseline registry contract:
// Register adds an entry; Lookup retrieves it; Has reports
// membership. The registry is process-lifetime, so each test uses
// a unique ScryfallID to avoid collisions with sibling tests.
func TestRegisterAndLookup(t *testing.T) {
	const id = "test-registry-baseline"
	registerForTest(t, Spec{OracleID: id, Name: "Baseline"})
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
	registerForTest(t, Spec{OracleID: id, Name: "First"})
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
	registerForTest(t, Spec{
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
	registerForTest(t, Spec{OracleID: id, Name: "SnapshotProbe"})
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

// TestRegisterExileSelfOffTheGraveyardPanics pins #1221's boot check:
// scavenge's and embalm's "Exile this card from YOUR GRAVEYARD" can
// only be paid by a card in a graveyard, so an ability that declares
// the component without declaring the zone could never be activated
// — and would look complete on the catalog page while refusing every
// activation. The symmetric check DiscardSelf has had since #660.
func TestRegisterExileSelfOffTheGraveyardPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("an exile-this cost off the graveyard did not panic")
		}
	}()
	Register(Spec{
		OracleID: "test-registry-exile-self-zone",
		Name:     "Misplaced Scavenge",
		Activated: []ActivatedAbility{{
			Label: "{1}, Exile this card: do nothing",
			Cost:  Plus(ManaCost("{1}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneHand},
		}},
	})
}

// And the other side of the same bound: with the graveyard declared
// it registers.
func TestRegisterExileSelfFromTheGraveyardIsFine(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: "test-registry-exile-self-ok",
		Name:     "Proper Scavenge",
		Activated: []ActivatedAbility{{
			Label: "{1}, Exile this card from your graveyard: do nothing",
			Cost:  Plus(ManaCost("{1}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
		}},
	})
	if !Has("test-registry-exile-self-ok") {
		t.Error("a graveyard exile-this ability did not register")
	}
}

// TestRegisterStaticFromAnUnwalkedZonePanics pins #1221's other
// boot check: a static may declare the zone it functions from, and
// the layer gather walks the graveyard and nothing else. A zone it
// does not walk would be a declaration the engine silently ignored —
// the card would register, look complete on the catalog page, and
// never apply. Same treatment #925 gives a trigger zone.
func TestRegisterStaticFromAnUnwalkedZonePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("a static declaring an unwalked zone did not panic")
		}
	}()
	Register(Spec{
		OracleID: "test-registry-static-zone",
		Name:     "Misplaced Incarnation",
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Zones: []game.ZoneKind{game.ZoneExile},
		}},
	})
}

// And the other side: the graveyard registers.
func TestRegisterStaticFromTheGraveyardIsFine(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: "test-registry-static-zone-ok",
		Name:     "Proper Incarnation",
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Zones: []game.ZoneKind{game.ZoneGraveyard},
		}},
	})
	if !Has("test-registry-static-zone-ok") {
		t.Error("a graveyard static did not register")
	}
}
