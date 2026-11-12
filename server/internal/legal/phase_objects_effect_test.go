package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phase_objects_effect_test.go — #5685, ADR 0188 Decision 14. A
// requirement and a restriction waiting for one added combat over named
// creatures (Illusionist's Gambit) reach the enumerator through the
// paths every other requirement and restriction uses: nothing is owed or
// withheld in a newer combat that runs first (CR 500.8), and in the
// bound combat the pass is withheld, the protected player and their
// planeswalker are not offered, every offered move is accepted and the
// withheld ones really are refused.

func TestThoseCreaturesThatCombatMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, protected := g.Seats[s], g.Seats[(s+1)%4]
	clearHand(me)
	bound := enteredCard(g, me, game.Card{Name: "Bound Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	free := enteredCard(g, me, game.Card{Name: "Free Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].SummonedThisTurn = false
	}
	walker := battlefieldCard(g, protected, game.Card{
		Name: "Protected Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	advanceTo(t, g, game.StepBeginCombat)
	g.WithWriteLock(func() {
		ids := g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
		if len(ids) != 1 || !g.DuringCombatPhaseObjectsEffectForEffect(ids[0], uuid.Nil, []uuid.UUID{bound}, protected.ID,
			[]game.Mod{game.AddAttackRequirementMod(uuid.Nil), game.CantAttackPlayerMod(protected.ID)}, "test — those creatures") {
			t.Fatalf("setup: ids %v, or the binding was refused", ids)
		}
		// A newer combat, added afterwards, runs first.
		g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})

	advanceTo(t, g, game.StepEndCombat)
	advanceTo(t, g, game.StepDeclareAttackers)
	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 1 || attacksBy(moves, bound) != 4 {
		t.Fatalf("the newer combat: pass %d, attacks by the named creature %d; want the pass and all four targets: %v",
			countKind(moves, legal.KindPass), attacksBy(moves, bound), labels(moves))
	}

	advanceTo(t, g, game.StepEndCombat)
	advanceTo(t, g, game.StepDeclareAttackers)
	moves = legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 0 {
		t.Fatalf("pass offered while the named creature owes an attack: %v", labels(moves))
	}
	offered := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != bound {
			continue
		}
		ap := decodeAttack(t, m)
		if ap.Target == protected.ID.String() || ap.Target == walker.String() {
			t.Errorf("offered the named creature at the protected player or their planeswalker: %q", m.Label)
		}
		offered++
	}
	if offered != 2 {
		t.Fatalf("offered %d attacks by the named creature, want the two other opponents: %v", offered, labels(moves))
	}
	if n := attacksBy(moves, free); n != 4 {
		t.Errorf("offered %d attacks by the creature that was not named, want all four targets: %v", n, labels(moves))
	}
	for _, refused := range []uuid.UUID{protected.ID, walker} {
		if err := g.Clone().DeclareAttacker(bound, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}
