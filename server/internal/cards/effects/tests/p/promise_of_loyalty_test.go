package testsp

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// promise_of_loyalty_test.go — Promise of Loyalty (#3400): each player
// keeps one creature, and the kept creatures can't attack the caster or
// their planeswalkers while they hold a vow counter.

const promiseOfLoyaltyOracle = "e1d294e8-67b1-4a4a-86b8-f98f28a327d9"

func TestPromiseOfLoyaltyIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(promiseOfLoyaltyOracle)
	if !ok || spec.Name != "Promise of Loyalty" {
		t.Fatalf("not registered: %+v", spec)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

func TestPromiseOfLoyaltyKeepsOneVowedCreaturePerPlayerAndBarsAttacksOnTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	third := g.Seats[(seat+2)%len(g.Seats)]
	myKept := effectstest.TdCreature(g, me.ID, "My Kept", 2, 2)
	myLost := effectstest.TdCreature(g, me.ID, "My Lost", 1, 1)
	theirKept := effectstest.TdCreature(g, opp.ID, "Their Kept", 3, 3)
	theirLost := effectstest.TdCreature(g, opp.ID, "Their Lost", 1, 1)
	theirLost2 := effectstest.TdCreature(g, opp.ID, "Their Lost Too", 1, 1)
	thirdKept := effectstest.TdCreature(g, third.ID, "Third Kept", 2, 2)
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Promise of Loyalty", TypeLine: "Sorcery",
		OracleID: promiseOfLoyaltyOracle, ManaCost: "{4}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)

	// One prompt at a time, caster first; nothing is sacrificed until the
	// last has answered.
	effectstest.AnswerKeep(t, g, me.ID, myKept)
	if !g.Battlefield.Contains(myLost) {
		t.Fatal("a creature was sacrificed before every player had chosen")
	}
	effectstest.AnswerKeep(t, g, opp.ID, theirKept)
	effectstest.AnswerKeep(t, g, third.ID, thirdKept)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{myLost, theirLost, theirLost2} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s should have been sacrificed", id)
		}
	}
	for _, id := range []uuid.UUID{myKept, theirKept, thirdKept} {
		if !g.Battlefield.Contains(id) || effectstest.CountersOn(g, id, "vow") != 1 {
			t.Errorf("%s should be kept with one vow counter", id)
		}
	}

	// Their turn: the kept creature may not attack the caster or the
	// caster's planeswalker, but may attack anyone else.
	oppSeat := (seat + 1) % len(g.Seats)
	effectstest.AdvanceToDeclareAttackersOf(t, g, oppSeat)
	for _, target := range []uuid.UUID{me.ID, walker} {
		err := g.Clone().DeclareAttacker(theirKept, target)
		var refused *game.AttackTargetRestrictionError
		if !errors.As(err, &refused) || !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Fatalf("attacking %s: err = %v, want an attack-target restriction refusal", target, err)
		}
	}
	if err := g.Clone().DeclareAttacker(theirKept, third.ID); err != nil {
		t.Fatalf("attacking another player: %v", err)
	}

	// The enumerator agrees.
	got := effectstest.BloodthirsterTargets(t, g, opp.ID, theirKept)
	if got[me.ID.String()] || got[walker.String()] || !got[third.ID.String()] {
		t.Errorf("enumerated targets %v", got)
	}

	// CR 611.2b: with the last vow counter gone, it ends for good.
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(theirKept, "vow", -1)
		g.RecomputeLayersIfStaleLocked()
	})
	if err := g.Clone().DeclareAttacker(theirKept, me.ID); err != nil {
		t.Fatalf("attacking the caster with no vow counter: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirKept, "vow", 1) })
	if err := g.Clone().DeclareAttacker(theirKept, me.ID); err != nil {
		t.Fatalf("a new vow counter must not bring the restriction back: %v", err)
	}
}

func TestPromiseOfLoyaltySurvivesASnapshotRestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.TdCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.TdCreature(g, opp.ID, "Theirs", 2, 2)

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Promise of Loyalty", TypeLine: "Sorcery",
		OracleID: promiseOfLoyaltyOracle, ManaCost: "{4}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerKeep(t, g, me.ID, mine)
	effectstest.AnswerKeep(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	var found bool
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModCantAttackPlayer && m.Player == me.ID {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no cantAttackPlayer record naming the caster")
	}
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, restored, (seat+1)%len(restored.Seats))
	if err := restored.Clone().DeclareAttacker(theirs, me.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("after a restore, attacking the caster: err = %v, want ErrIllegalAttackTarget", err)
	}
}
