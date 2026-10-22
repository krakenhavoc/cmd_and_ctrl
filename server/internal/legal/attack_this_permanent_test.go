package legal_test

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestAttackThisPermanentMovesAgreeWithTheEngine — #2567: a creature that
// must attack one planeswalker (Gideon Jura's requirement) is offered
// that one attack, as AlwaysLegal; its other attacks and the pass are
// withheld, and the engine refuses each of them.
func TestAttackThisPermanentMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	clearHand(me)
	bear := freshCreature(g, me, "Bound Bear")
	walker := battlefieldCard(g, owner, game.Card{
		Name: "Gideon Jura", TypeLine: "Legendary Planeswalker — Gideon",
		Counters: map[string]int{game.CounterLoyalty: 6},
	})
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(walker, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.AttackPermanentIfAbleMod(game.ObjectRef{ID: walker}, 0)},
			game.IndefiniteDuration(), "test — attacks Gideon Jura if able") {
			t.Fatal("registered nothing")
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 0 {
		t.Fatalf("pass offered while the creature owes its attack: %v", labels(moves))
	}
	offered := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != bear {
			continue
		}
		offered++
		if ap := decodeAttack(t, m); ap.Target != walker.String() {
			t.Errorf("offered an attack at %s, want only the planeswalker", ap.Target)
		}
		if !m.AlwaysLegal {
			t.Errorf("the attack that answers the requirement is not AlwaysLegal: %q", m.Label)
		}
	}
	if offered != 1 {
		t.Fatalf("offered %d attacks for the bound creature, want 1: %v", offered, labels(moves))
	}
	if err := g.Clone().DeclareAttacker(bear, owner.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("engine accepted the withheld attack at the walker's controller: %v", err)
	}
	if err := g.Clone().PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("engine accepted the withheld pass: %v", err)
	}
	if err := g.DeclareAttacker(bear, walker); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, me.ID)
	if countKind(moves, legal.KindPass) != 1 {
		t.Errorf("pass not offered once the requirement is obeyed: %v", labels(moves))
	}
}
