package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// own_ability_removal_test.go — #1859: the loseOwnAbility mod. One
// replacement, triggered or activated row of an object's own definition
// is switched off in layer 6, for the record's duration, and nothing else
// is.

const (
	oarCard   = "own-ability-fixture/card"
	oarBundle = "own-ability-fixture/granted-trigger"
)

func oarTrigger(key string) TriggeredAbility {
	return TriggeredAbility{Key: key, Watches: []EventKind{EventETB}}
}

// stubOwnAbilityCatalog files a card with two rows in each of the three
// slots, and a bundle that grants one trigger.
func stubOwnAbilityCatalog(t *testing.T) {
	t.Helper()
	defs := map[string]*CardDef{
		oarCard: {
			Replacements: []ReplacementEffect{{Label: "rep-0"}, {Label: "rep-1"}},
			Triggered:    []TriggeredAbility{oarTrigger("trig-0"), oarTrigger("trig-1")},
			Activated: []ActivatedAbilityShape{
				{Label: "act-0"},
				{Label: "act-1"},
			},
		},
		GrantKey(oarBundle): {
			Triggered: []TriggeredAbility{oarTrigger("trig-0")},
		},
	}
	IdentifyCatalogRows(oarCard, defs[oarCard])
	IdentifyCatalogRows(GrantKey(oarBundle), defs[GrantKey(oarBundle)])
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

func pushOARCreature(g *Game, owner uuid.UUID) uuid.UUID {
	id := pushScopedTestCreature(g, owner, 2, 2)
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		c.OracleID = oarCard
		g.layerVersion.Add(1)
	})
	return id
}

func oarRegister(t *testing.T, g *Game, id uuid.UUID, d Duration, mods ...Mod) {
	t.Helper()
	sgRegister(t, g, id, id, mods, func() Duration { return d })
}

// oarRows reads what each slot holds on the permanent right now.
func oarRows(t *testing.T, g *Game, id uuid.UUID) (reps, trigs, acts []string, actRefs []string) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("card %s not on battlefield", id)
		}
		for _, r := range CatalogReplacements(CatalogAbilityKey(*c)) {
			reps = append(reps, r.Label)
		}
		for _, tr := range TriggersForCard(*c) {
			trigs = append(trigs, tr.Key)
		}
		rows, origins := ActivatedAbilitiesWithOrigins(*c)
		for i, a := range rows {
			acts = append(acts, a.Label)
			actRefs = append(actRefs, origins.Ref(i))
		}
	})
	return
}

func oarSame(a, b []string) bool { return strings.Join(a, ",") == strings.Join(b, ",") }

func TestLoseOwnAbilitySwitchesOffExactlyOneRowPerSlot(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		mod                   Mod
		reps, trigs, acts, rf string
	}{
		{"nothing removed", Mod{}, "rep-0,rep-1", "trig-0,trig-1", "act-0,act-1", "own:0,own:1"},
		{"a replacement", LoseOwnAbilityMod(AbilitySlotReplacement, 0), "rep-1", "trig-0,trig-1", "act-0,act-1", "own:0,own:1"},
		{"a triggered", LoseOwnAbilityMod(AbilitySlotTriggered, 1), "rep-0,rep-1", "trig-0", "act-0,act-1", "own:0,own:1"},
		{"an activated, refs held", LoseOwnAbilityMod(AbilitySlotActivated, 0), "rep-0,rep-1", "trig-0,trig-1", "act-1", "own:1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stubOwnAbilityCatalog(t)
			g := newActiveGame(t)
			id := pushOARCreature(g, g.Seats[0].ID)
			if tc.mod.Kind != "" {
				oarRegister(t, g, id, IndefiniteDuration(), tc.mod)
			}
			reps, trigs, acts, refs := oarRows(t, g, id)
			if !oarSame(reps, strings.Split(tc.reps, ",")) || !oarSame(trigs, strings.Split(tc.trigs, ",")) ||
				!oarSame(acts, strings.Split(tc.acts, ",")) || !oarSame(refs, strings.Split(tc.rf, ",")) {
				t.Errorf("rows = %v / %v / %v (%v)", reps, trigs, acts, refs)
			}
		})
	}
}

// CR 611.2a: until end of turn, and then it is back.
func TestLoseOwnAbilityEndsAtCleanup(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	var eot Duration
	g.WithWriteLock(func() { eot = g.UntilEndOfTurnDuration() })
	oarRegister(t, g, id, eot, LoseOwnAbilityMod(AbilitySlotReplacement, 0))
	if reps, _, _, _ := oarRows(t, g, id); !oarSame(reps, []string{"rep-1"}) {
		t.Fatalf("setup: replacements = %v", reps)
	}
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if reps, _, _, _ := oarRows(t, g, id); !oarSame(reps, []string{"rep-0", "rep-1"}) {
		t.Errorf("after cleanup replacements = %v, want both back", reps)
	}
}

// CR 613.6 / 613.1f: a grant with a LATER timestamp of the same text is a
// different row and is not removed; it is the object's own row that goes.
func TestALaterGrantOfTheSameTextSurvivesTheRemoval(t *testing.T) {
	stubOwnAbilityCatalog(t)
	tick := int64(1_000_000)
	defer SetClockForTest(func() int64 { tick += 1000; return tick })()
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	ind := IndefiniteDuration()
	oarRegister(t, g, id, ind, LoseOwnAbilityMod(AbilitySlotTriggered, 0))
	oarRegister(t, g, id, ind, GrantAbilitiesMod(oarBundle))
	_, trigs, _, _ := oarRows(t, g, id)
	if !oarSame(trigs, []string{"trig-1", "trig-0"}) {
		t.Errorf("triggers = %v, want the own trig-1 and the GRANTED trig-0 (own trig-0 is gone)", trigs)
	}
}

// loseAllAbilities dominates: with every own ability gone there is no row
// left to name, and the composite key stays empty.
func TestLoseAllAbilitiesDominatesARowRemoval(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	oarRegister(t, g, id, IndefiniteDuration(), LoseOwnAbilityMod(AbilitySlotReplacement, 0), LoseAllAbilitiesMod())
	reps, trigs, acts, _ := oarRows(t, g, id)
	if len(reps)+len(trigs)+len(acts) != 0 {
		t.Errorf("rows = %v %v %v, want none", reps, trigs, acts)
	}
}

// CR 400.7: a new object after a flicker is not the one the record named.
func TestARowRemovalDoesNotFollowItsObjectThroughAFlicker(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	oarRegister(t, g, id, IndefiniteDuration(), LoseOwnAbilityMod(AbilitySlotReplacement, 0))
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		c.EnteredBattlefieldAt += 1_000_000
		g.layerVersion.Add(1)
	})
	if reps, _, _, _ := oarRows(t, g, id); !oarSame(reps, []string{"rep-0", "rep-1"}) {
		t.Errorf("the new object has replacements %v, want both", reps)
	}
}

// A copy reads PrintedValues, never the layered result (CR 707.2), so the
// removal is not copiable by construction.
func TestRemovedOwnAbilitiesAreNotCopiable(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	oarRegister(t, g, id, IndefiniteDuration(), LoseOwnAbilityMod(AbilitySlotReplacement, 0))
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.battlefieldCardLocked(id)
		v := CopiableValuesOf(*c)
		if got := CatalogReplacements(v.OracleID); len(got) != 2 {
			t.Errorf("a copy has %d replacement rows, want the printed two", len(got))
		}
	})
}

// The dies harvest reads the snapshot, not the live card (CR 603.10a).
func TestTheLastKnownKeyKeepsTheRemoval(t *testing.T) {
	stubOwnAbilityCatalog(t)
	lki := Characteristic{RemovedOwnAbilities: []OwnAbilityRemoval{{Slot: AbilitySlotTriggered, Row: 0}}}
	key := AbilityKeyFromLKI(Card{OracleID: oarCard}, lki)
	got := CatalogTriggers(key)
	if len(got) != 1 || got[0].Key != "trig-1" {
		t.Errorf("last-known triggers = %v, want only trig-1", got)
	}
}

func TestARowRemovalIsARestorePoint(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	oarRegister(t, g, id, IndefiniteDuration(), LoseOwnAbilityMod(AbilitySlotReplacement, 1))
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"kind":"loseOwnAbility"`) {
		t.Errorf("the record is not data on disk:\n%s", raw)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if reps, _, _, _ := oarRows(t, restored, id); !oarSame(reps, []string{"rep-0"}) {
		t.Errorf("restored replacements = %v, want only rep-0", reps)
	}
	// Undo clones the game; the clone carries the record too.
	clone := g.Clone()
	if reps, _, _, _ := oarRows(t, clone, id); !oarSame(reps, []string{"rep-0"}) {
		t.Errorf("cloned replacements = %v, want only rep-0", reps)
	}
}

// The view lists abilities from the same accessors, so a removed row is
// not on the tile while it is gone (#2219's chips).
func TestARemovedRowLeavesTheAbilityChips(t *testing.T) {
	stubOwnAbilityCatalog(t)
	g := newActiveGame(t)
	id := pushOARCreature(g, g.Seats[0].ID)
	count := func(kind AbilityRowKind) int {
		n := 0
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c, _ := g.battlefieldCardLocked(id)
			for _, r := range AbilityRowsOf(*c) {
				if r.Kind == kind {
					n++
				}
			}
		})
		return n
	}
	before := [3]int{count(AbilityRowTriggered), count(AbilityRowStatic), count(AbilityRowActivated)}
	oarRegister(t, g, id, IndefiniteDuration(),
		LoseOwnAbilityMod(AbilitySlotReplacement, 0), LoseOwnAbilityMod(AbilitySlotTriggered, 0), LoseOwnAbilityMod(AbilitySlotActivated, 0))
	after := [3]int{count(AbilityRowTriggered), count(AbilityRowStatic), count(AbilityRowActivated)}
	for i := range before {
		if after[i] != before[i]-1 {
			t.Errorf("chip %d: %d before, %d after, want one fewer", i, before[i], after[i])
		}
	}
}

func TestARowRemovalNamingAnUnsupportedSlotIsRefused(t *testing.T) {
	for _, slot := range []string{"static", "mana", "", "bogus"} {
		if loseOwnAbilityModProblem(LoseOwnAbilityMod(slot, 0)) == "" {
			t.Errorf("slot %q was accepted", slot)
		}
	}
	if loseOwnAbilityModProblem(LoseOwnAbilityMod(AbilitySlotTriggered, -1)) == "" {
		t.Error("a negative row was accepted")
	}
	if loseOwnAbilityModProblem(Mod{Kind: ModAddKeywords, Slot: AbilitySlotTriggered}) == "" {
		t.Error("a slot on another kind was accepted")
	}
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 1)
	defer func() {
		if recover() == nil {
			t.Error("registering a static-slot removal did not panic")
		}
	}()
	oarRegister(t, g, id, IndefiniteDuration(), LoseOwnAbilityMod("static", 0))
}
