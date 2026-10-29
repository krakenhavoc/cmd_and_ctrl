package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func readObjectTokenCopy(g *Game, ref ObjectRef) (Card, bool) {
	var token Card
	var ok bool
	g.WithWriteLock(func() { token, ok = g.TokenCopyTemplateForObjectForEffect(ref) })
	return token, ok
}

func TestTokenCopyOfDepartedObjectPreservesItsCopyEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	src := bearsOnBattlefield(g, me, "Grizzly Bears")
	clone := bearsOnBattlefield(g, me, "Clone")
	g.WithWriteLock(func() {
		original := findBattlefieldCard(g, src)
		copier := findBattlefieldCard(g, clone)
		values := CopiableValuesOf(*original)
		values.GrantedAbilities = []string{"test/copied-ability"}
		copier.applyCopy(values, *original)
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(clone), []Mod{AddKeywordsMod("flying")}, g.UntilEndOfTurnDuration(), "Noncopiable flying")
		_ = g.AddCounterForEffect(clone, CounterPlusOne, 3)
	})
	ref := refOf(t, g, clone)
	live, ok := readObjectTokenCopy(g, ref)
	if !ok || live.Name != "Grizzly Bears" || !live.IsToken() || live.Power != 2 || live.IsCommander || len(live.Counters) != 0 {
		t.Fatalf("live copy = %+v, %v; want only copiable Grizzly Bears values", live, ok)
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(clone); err != nil {
			t.Fatal(err)
		}
	})
	g.WithWriteLock(func() {
		hand, ok := g.LookupCardForEffect(clone)
		if !ok || hand.Name != "Clone" || !g.Seats[0].Hand.Contains(clone) {
			t.Fatalf("departed card should revert to Clone, got %+v, %v", hand, ok)
		}
	})
	undo := g.Clone()
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snapshot GameSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	restored, err := snapshot.Restore()
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []*Game{g, undo, restored} {
		copy, ok := readObjectTokenCopy(state, ref)
		if !ok || copy.Name != "Grizzly Bears" || copy.OracleID != "oracle-Grizzly Bears" || copy.Power != 2 || copy.Keywords[0] != "trample" || len(copy.Keywords) != 1 || len(copy.GrantedAbilities) != 1 {
			t.Fatalf("departed object's token copy = %+v, %v", copy, ok)
		}
	}
	// A new object made from the same card must never supply this trigger's copy.
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me}, ZoneRef{Kind: ZoneBattlefield}, clone); err != nil {
		t.Fatal(err)
	}
	copy, ok := readObjectTokenCopy(g, ref)
	if !ok || copy.Name != "Grizzly Bears" {
		t.Fatalf("returned Clone overwrote its former object's copy: %+v, %v", copy, ok)
	}
	// The undo record owns its PrintedValues and their slices.
	g.lastKnownPermanents[clone][0].Copiable.Keywords[0] = "haste"
	copy, _ = readObjectTokenCopy(undo, ref)
	if copy.Keywords[0] != "trample" {
		t.Fatal("live LKI mutated the undo point's copiable values")
	}
}

func TestSimultaneousTokenEntryPauseAndUndo(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kismet": {entersTappedForOthers("Kismet: enters tapped")},
		"thalia": {entersTappedForOthers("Thalia: enters tapped")},
	})
	pushReplacementSource(g, "kismet", me)
	pushReplacementSource(g, "thalia", me)
	ran := 0
	g.WithWriteLock(func() {
		spec := TokenCreation{Controller: them, Together: true,
			Groups: []TokenGroup{{Template: goblinToken(), Count: 2, Entry: TokenEntryOptions{Counters: map[string]int{CounterPlusOne: 1}}}},
		}
		// #4361 S10: tokens entering together carry their continuation as
		// data on the batch (#4361 S12 deleted the closure door).
		if err := g.CreateTokensKeyedForEffect(spec, hookTail(func(_ *Game, o TailOutcome) error {
			ran++
			if len(o.IDs) != 2 {
				t.Errorf("continuation got %d tokens, want 2", len(o.IDs))
			}
			return nil
		})); err != nil {
			t.Fatal(err)
		}
	})
	first := onlyReplacementOrderPrompt(t, g, them)
	if err := g.ResolveReplacementOrder(first.ID, them, first.ReplacementEffectIDs); err != nil {
		t.Fatal(err)
	}
	if tokensNamed(g, "Goblin") != 0 || ran != 0 {
		t.Fatal("first token landed before the simultaneous entry's second window settled")
	}
	paused := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			ran = 0
			g.WithWriteLock(func() { g.RestoreFrom(paused) })
		}
		second := onlyReplacementOrderPrompt(t, g, them)
		if err := g.ResolveReplacementOrder(second.ID, them, second.ReplacementEffectIDs); err != nil {
			t.Fatal(err)
		}
		if tokensNamed(g, "Goblin") != 2 || ran != 1 {
			t.Fatalf("run %d: %d Goblins, %d continuation runs", run, tokensNamed(g, "Goblin"), ran)
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Goblin" && (!c.Tapped || c.Counters[CounterPlusOne] != 1) {
				t.Fatal("simultaneous token entry lost its tapped status or entry counter")
			}
		}
	}
}
