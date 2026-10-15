package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// second_combat_test.go — ADR 0188 Decision 10, #3566: the agreement test
// ADR 0059 Decision 13 left open. A combat an effect added (Relentless
// Assault, Combat Celebrant) has its own declare attackers step, and the
// enumerator must offer the active seat attacks there exactly as in the
// first one, every one of them accepted by the dispatcher.

// TestSecondCombatOffersAttacksAndDispatchAcceptsThem walks a turn whose
// first combat attacks with one creature and whose added combat attacks
// with another, checking the enumerator at both declare attackers steps.
func TestSecondCombatOffersAttacksAndDispatchAcceptsThem(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	first := battlefieldCard(g, active, creature("First Wave", "{G}", 2, 2))
	second := battlefieldCard(g, active, creature("Second Wave", "{G}", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)

	if g.Turn.PhaseOrdinal > 1 {
		t.Fatalf("the first combat is phase ordinal %d, want 1", g.Turn.PhaseOrdinal)
	}
	moves := legal.EnumerateFor(g, active.ID)
	if countKind(moves, legal.KindAttack) == 0 {
		t.Fatalf("the first combat offers no attack: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// Add the combat while this one is under way, then attack with one
	// creature now and keep the other for the added combat.
	var ids []int
	g.WithWriteLock(func() {
		ids = g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	if len(ids) != 1 {
		t.Fatalf("AddPhasesForEffect added %v, want one combat", ids)
	}
	if err := g.DeclareAttacker(first, def.ID); err != nil {
		t.Fatal(err)
	}
	walkToSecondDeclareAttackers(t, g)

	if g.Turn.PhaseOrdinal != 2 {
		t.Fatalf("phase ordinal %d in the added combat, want 2", g.Turn.PhaseOrdinal)
	}
	if !holdsPriority(g, active.ID) {
		t.Fatalf("the active seat should hold priority as the added declare attackers begins (holder %d)", g.Turn.PriorityHolder)
	}
	moves = legal.EnumerateFor(g, active.ID)
	if countKind(moves, legal.KindAttack) == 0 {
		t.Fatalf("the added combat offers no attack: %v", labels(moves))
	}
	// The creature that has not attacked yet is among them, and every
	// offered move is one the dispatcher takes.
	offered := false
	for _, m := range moves {
		if m.Kind == legal.KindAttack && string(m.Params) != "" && strings.Contains(string(m.Params), second.String()) {
			offered = true
		}
	}
	if !offered {
		t.Errorf("the creature held back for the added combat is not offered an attack: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// walkToSecondDeclareAttackers advances step by step, passing priority
// where a step holds it, until the added combat's declare attackers.
func walkToSecondDeclareAttackers(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 60; i++ {
		if g.Turn.Step == game.StepDeclareAttackers && g.Turn.PhaseOrdinal == 2 {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep from %s: %v", g.Turn.Step, err)
		}
	}
	t.Fatalf("never reached the added combat's declare attackers (at %s, ordinal %d)", g.Turn.Step, g.Turn.PhaseOrdinal)
}
