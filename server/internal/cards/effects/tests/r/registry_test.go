package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRegisterAndLookup covers the baseline registry contract:
// Register adds an entry; Lookup retrieves it; Has reports
// membership. The registry is process-lifetime, so each test uses
// a unique ScryfallID to avoid collisions with sibling tests.
func TestRegisterAndLookup(t *testing.T) {
	const id = "test-registry-baseline"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: id, Name: "Baseline"})
	spec, ok := effects.Lookup(id)
	if !ok {
		t.Fatalf("Lookup(%q): ok=false, want true", id)
	}
	if spec.Name != "Baseline" {
		t.Errorf("Lookup(%q).Name: got %q, want %q", id, spec.Name, "Baseline")
	}
	if !effects.Has(id) {
		t.Errorf("Has(%q): false, want true", id)
	}
}

// TestLookupMiss covers the non-catalog path: an unknown ID
// returns ok=false and Has returns false. The resolution path
// keys its auto-vs-manual decision off this signal.
func TestLookupMiss(t *testing.T) {
	const id = "test-registry-nonexistent-1234"
	if _, ok := effects.Lookup(id); ok {
		t.Errorf("Lookup(%q): ok=true, want false (non-registered ID)", id)
	}
	if effects.Has(id) {
		t.Errorf("Has(%q): true, want false", id)
	}
}

// TestRegisterDuplicatePanics pins the copy-paste catch: two Specs
// with the same ScryfallID panic at registration, surfacing the
// bug at server boot rather than silently letting one win.
func TestRegisterDuplicatePanics(t *testing.T) {
	const id = "test-registry-duplicate"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: id, Name: "First"})
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("duplicate Register did not panic")
		}
	}()
	effects.Register(effects.Spec{OracleID: id, Name: "Second"})
}

// TestRegisterEmptyIDPanics pins the sanity check: a Spec with no
// key would shadow Lookup("") and mask caller bugs.
func TestRegisterEmptyIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("empty-ID Register did not panic")
		}
	}()
	effects.Register(effects.Spec{Name: "MissingID"})
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
	effects.Register(effects.Spec{
		OracleID:     "test-registry-replacement-budget",
		Name:         "Slot Hog",
		Replacements: make([]game.ReplacementEffect, game.MaxCatalogReplacementSlots+1),
	})
}

// TestRegisterAtTheReplacementBudgetIsFine is the other side of the
// bound: the budget is inclusive, so a Spec exactly at it registers.
func TestRegisterAtTheReplacementBudgetIsFine(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     "test-registry-replacement-budget-exact",
		Name:         "Slot Hog Jr",
		Replacements: make([]game.ReplacementEffect, game.MaxCatalogReplacementSlots),
	})
	if !effects.Has("test-registry-replacement-budget-exact") {
		t.Errorf("a Spec at the slot budget did not register")
	}
}

// TestAllReturnsSnapshot proves All returns a fresh slice —
// mutating it does not leak back into the registry.
func TestAllReturnsSnapshot(t *testing.T) {
	const id = "test-registry-all-snapshot"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: id, Name: "SnapshotProbe"})
	snap := effects.All()
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
	before := len(effects.All())
	snap[0] = effects.Spec{} // mutate caller-owned slice
	after := len(effects.All())
	if before != after {
		t.Errorf("mutating All() leaked back: before=%d, after=%d", before, after)
	}
}

// TestRegisterExileSelfOffTheGraveyardPanics pins #1221's boot check,
// as #1404 and #3520 widened it: the exile-this component is paid from
// the graveyard, the battlefield or the hand, so an ability that
// functions from EXILE could never be activated — and would look complete on the
// catalog page while refusing every activation. The symmetric check
// DiscardSelf has had since #660.
func TestRegisterExileSelfOffTheGraveyardPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("an exile-this cost off the graveyard did not panic")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-registry-exile-self-zone",
		Name:     "Misplaced Scavenge",
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Exile this card: do nothing",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.ExileThis()),
			Zones: []game.ZoneKind{game.ZoneExile},
		}},
	})
}

// And the other side of the same bound: with the graveyard declared
// it registers.
func TestRegisterExileSelfFromTheGraveyardIsFine(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "test-registry-exile-self-ok",
		Name:     "Proper Scavenge",
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Exile this card from your graveyard: do nothing",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
		}},
	})
	if !effects.Has("test-registry-exile-self-ok") {
		t.Error("a graveyard exile-this ability did not register")
	}
}

// #1404: and the battlefield, which is the default zone — Perpetual
// Timepiece's "{2}, Exile this artifact:" registers with no Zones at
// all. Before #1404 this panicked with "does not function from the
// graveyard"; the hand test above still panics, because no CR 602
// ability prints "exile this card from your hand".
func TestRegisterExileSelfFromTheBattlefieldIsFine(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "test-registry-exile-self-battlefield",
		Name:     "Proper Timepiece",
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Exile this artifact: do nothing",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.ExileThis()),
		}},
	})
	if !effects.Has("test-registry-exile-self-battlefield") {
		t.Error("a battlefield exile-this ability did not register")
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
	effects.Register(effects.Spec{
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
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "test-registry-static-zone-ok",
		Name:     "Proper Incarnation",
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Zones: []game.ZoneKind{game.ZoneGraveyard},
		}},
	})
	if !effects.Has("test-registry-static-zone-ok") {
		t.Error("a graveyard static did not register")
	}
}
