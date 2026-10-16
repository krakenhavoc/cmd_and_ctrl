package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pinned_retarget_slot_test.go — #3860's card proofs: Spellskite and
// Mizzium Meddler ask WHICH instance of "target" changes when the same
// creature fills several. The engine half is
// game/retarget_pinned_test.go.

// ployAtOneCreatureTwice casts an Azorius Ploy ("target creature … target
// creature", no "another") aimed at `victim` in BOTH slots: two instances
// of the word "target" that chose one creature (CR 601.2c). It is the
// card #3860 needs, because the two sentences do different things -- the
// first stops `victim` dealing combat damage, the second stops combat
// damage being dealt to it -- so which slot becomes Spellskite matters.
func ployAtOneCreatureTwice(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	ref0 := game.TargetRef{Kind: game.TargetCard, ID: victim, Slot: 0}
	ref1 := game.TargetRef{Kind: game.TargetCard, ID: victim, Slot: 1}
	return castCatalogSpell(t, g, "Azorius Ploy", "Instant", pr7bAzoriusPloy, []game.TargetRef{ref0, ref1})
}

// #3860: with the same creature chosen for both instances of "target",
// Spellskite's controller says which one changes -- the ruling the
// first cut guessed at (the earlier slot) -- instead of the engine
// choosing for them.
func TestSpellskiteAsksWhichInstanceOfTargetChangesToIt(t *testing.T) {
	for _, tc := range []struct {
		name string
		pick int
		slot int // the slot that becomes Spellskite
	}{
		{"the first instance", 0, 0},
		{"the second instance", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			skite := pushSpellskite(g, opp.ID)
			victim := pushCatalogPermanent(g, me.ID, "Victim", "Creature — Bear", "", false)
			ploy := ployAtOneCreatureTwice(t, g, victim)

			activateSpellskite(t, g, opp.ID, skite, ploy)
			passPriorityAroundTable(t, g)

			prompt := latestOptionPickFor(g, opp.ID)
			if prompt == nil {
				t.Fatal("Spellskite's controller was not asked which instance of target changes")
			}
			if len(prompt.PickOptions) != 2 {
				t.Fatalf("options = %+v, want one per instance of target", prompt.PickOptions)
			}
			if got := g.StackMeta[ploy].Targets; got[0].ID != victim || got[1].ID != victim {
				t.Fatalf("nothing moves before the answer: %v", got)
			}
			if err := g.ResolveOptionPick(prompt.ID, opp.ID, tc.pick); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			got := g.StackMeta[ploy].Targets
			for slot := range got {
				want := victim
				if slot == tc.slot {
					want = skite
				}
				if got[slot].ID != want {
					t.Errorf("slot %d targets %v, want %v (slot %d is the one that changes)", slot, got[slot].ID, want, tc.slot)
				}
			}
		})
	}
}

// Mizzium Meddler's "you may" is asked over the creature first (so the
// whole thing can be declined), and then, because that creature fills
// two instances of "target", which one changes.
func TestMizziumMeddlerAsksWhichInstanceOfTargetChanges(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := pushCatalogPermanent(g, me.ID, "Victim", "Creature — Bear", "", false)
	ploy := ployAtOneCreatureTwice(t, g, victim)
	meddler := castCatalogSpell(t, g, "Mizzium Meddler", "Creature — Vedalken Wizard", mizziumMeddlerOracle, nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == meddler {
				g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 1, 4
			}
		}
	})
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil && latestRetarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, ploy)
	}
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Mizzium Meddler's trigger asked nothing")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	slot := latestOptionPickFor(g, me.ID)
	if slot == nil || len(slot.PickOptions) != 2 {
		t.Fatalf("after naming the creature the slot is asked, got %+v", slot)
	}
	if err := g.ResolveOptionPick(slot.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	got := g.StackMeta[ploy].Targets
	if got[0].ID != victim || got[1].ID != meddler {
		t.Errorf("targets = [%v %v], want the second instance on the Meddler", got[0].ID, got[1].ID)
	}
}

func TestSpellskiteAndMizziumMeddlerAreComplete(t *testing.T) {
	for _, oracle := range []string{spellskiteOracle, mizziumMeddlerOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness = %v, want full (#3860)", spec.Name, spec.Completeness)
		}
	}
}
