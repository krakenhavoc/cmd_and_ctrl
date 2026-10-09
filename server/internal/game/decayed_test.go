package game

import (
	"testing"
)

// decayed_test.go — the engine half of #2650: the canonical token, its
// counter (CR 122.1b), "This creature can't block" and "When this
// creature attacks, sacrifice it at end of combat" (CR 702.147a).

func TestDecayedIsCumulativeAndCanonical(t *testing.T) {
	if kw, ok := CanonicalKeyword("Decayed"); !ok || kw != KeywordDecayed {
		t.Errorf("CanonicalKeyword(\"Decayed\") = %q, %v", kw, ok)
	}
	if !KeywordIsCumulative(KeywordDecayed) {
		t.Error("decayed is not cumulative; CR 113.2c makes each instance its own trigger")
	}
	if !IsKeywordCounter(CounterDecayed) || CounterDecayed != KeywordDecayed {
		t.Error("the decayed counter is not a keyword counter for the decayed token")
	}
	printed := Card{Name: "Zombie", TypeLine: "Creature — Zombie", Keywords: []string{KeywordDecayed}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordDecayed {
		t.Fatalf("a printed decayed with no catalog entry has triggers %+v, want one decayed trigger", trigs)
	}
	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d decayed triggers", n)
	}
}

func TestDecayedCantBlock(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushCombatant(t, g, me, "Zombie", 2, 2, KeywordDecayed)
	plain := pushCombatant(t, g, me, "Bear", 2, 2)
	g.WithWriteLock(func() { g.layerVersion.Add(1) })
	if scopedEffectChar(t, g, id).Restrictions&CantBlock == 0 {
		t.Error("a decayed creature can block")
	}
	if scopedEffectChar(t, g, plain).Restrictions&CantBlock != 0 {
		t.Fatal("a plain creature can't block")
	}
	// A decayed counter gives the keyword, and with it the restriction.
	addCounterForTest(t, g, plain, CounterDecayed, 1)
	if scopedEffectChar(t, g, plain).Restrictions&CantBlock == 0 {
		t.Error("a creature with a decayed counter can block")
	}
}

// decayedItems is every queued or stacked item carrying the given body.
func decayedItems(g *Game, body string) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Body == body {
				out = append(out, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Body == body {
				out = append(out, it)
			}
		}
	})
	return out
}

func TestDecayedIsSacrificedAtEndOfCombat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushCombatant(t, g, me, "Zombie", 2, 2, KeywordDecayed)
	bystander := pushCombatant(t, g, me, "Decayed Bystander", 2, 2, KeywordDecayed)
	declareAndLock(t, g, opp.ID, zombie)

	items := decayedItems(g, decayedScheduleKey)
	if len(items) != 1 {
		t.Fatalf("one decayed attacker queued %d triggers, want 1", len(items))
	}
	if it := items[0]; it.Params.Object.ID != zombie || it.Controller != me.ID {
		t.Errorf("trigger = %+v, want the attacker pinned under my control", it)
	}
	resolveQueuedTrigger(g)
	var queued int
	g.ReadSnapshot(func() {
		for _, dt := range g.DelayedTriggers {
			if dt.Body.key == decayedSacrificeKey && dt.At == StepEndCombat && dt.Params.Object.ID == zombie {
				queued++
			}
		}
	})
	if queued != 1 {
		t.Fatalf("%d end-of-combat sacrifices queued, want 1", queued)
	}
	if findCard(g, zombie) == nil {
		t.Fatal("the zombie was sacrificed before end of combat")
	}

	passUntilStep(t, g, StepPostcombatMain)
	if findCard(g, zombie) != nil || !inGraveyard(me, zombie) {
		t.Error("the decayed attacker was not sacrificed at end of combat")
	}
	if findCard(g, bystander) == nil {
		t.Error("a decayed creature that did not attack was sacrificed")
	}
}

// Each instance is its own trigger (CR 113.2c); the second sacrifice
// finds nothing and does nothing.
func TestTwoDecayedInstancesSacrificeOnce(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushCombatant(t, g, me, "Zombie", 2, 2, KeywordDecayed)
	addCounterForTest(t, g, zombie, CounterDecayed, 1)
	if n := countOf(scopedEffectChar(t, g, zombie).Abilities, KeywordDecayed); n != 2 {
		t.Fatalf("printed decayed and a decayed counter = %d instances, want 2", n)
	}
	declareAndLock(t, g, opp.ID, zombie)
	if n := len(decayedItems(g, decayedScheduleKey)); n != 2 {
		t.Fatalf("two decayed instances queued %d triggers, want 2", n)
	}
	passUntilStep(t, g, StepPostcombatMain)
	if !inGraveyard(me, zombie) {
		t.Error("the decayed attacker was not sacrificed")
	}
}

// CR 400.7: a creature that left and came back before end of combat is
// a new object, and the delayed trigger leaves it alone.
func TestDecayedSkipsANewObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushCombatant(t, g, me, "Zombie", 2, 2, KeywordDecayed)
	declareAndLock(t, g, opp.ID, zombie)
	resolveQueuedTrigger(g)
	g.WithWriteLock(func() { findBattlefieldCard(g, zombie).ObjectEpoch++ })
	passUntilStep(t, g, StepPostcombatMain)
	if findCard(g, zombie) == nil {
		t.Error("a new object was sacrificed by the old object's decayed trigger")
	}
}
