package legal_test

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_requirements_test.go — #1571, CR 508.1d. The enumerator and
// the verbs agree: every attack offered while a requirement is owed is
// accepted, the pass is withheld exactly while the engine would refuse
// it, the attacks that answer the requirement are AlwaysLegal (so a bot
// that declines is not left holding the table), and every withheld move
// really is refused.

// TestGoadedCreatureMovesAgreeWithTheEngine — goad in four seats.
func TestGoadedCreatureMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goader := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	bear := freshCreature(g, me, "Goaded Bear")
	if err := g.SetGoaded(bear, goader.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 0 {
		t.Fatalf("pass offered while the goaded creature owes an attack: %v", labels(moves))
	}
	required := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != bear {
			continue
		}
		ap := decodeAttack(t, m)
		if ap.Target == goader.ID.String() {
			t.Errorf("offered the goaded creature at its goader while other opponents are open: %q", m.Label)
		}
		if !m.AlwaysLegal {
			t.Errorf("an attack that answers the goad is not AlwaysLegal: %q", m.Label)
		}
		required++
	}
	if required != 2 {
		t.Fatalf("goaded creature offered %d attacks, want the two non-goader opponents: %v", required, labels(moves))
	}
	// The withheld moves are refused by the engine — the enumerator is
	// not simply silent on both sides.
	if err := g.Clone().DeclareAttacker(bear, goader.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("engine accepted the withheld attack at the goader: %v", err)
	}
	if err := g.Clone().PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("engine accepted the withheld pass: %v", err)
	}

	// Answer it; the pass comes back.
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 1 {
		t.Errorf("pass not offered once the requirement is obeyed: %v", labels(moves))
	}
}

// TestUngoadedTableOffersThePassAsBefore — nothing changes at a table
// with no requirement on it: the pass is offered and no attack is
// AlwaysLegal.
func TestUngoadedTableOffersThePassAsBefore(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	freshCreature(g, me, "Bear")
	advanceTo(t, g, game.StepDeclareAttackers)
	moves := legal.EnumerateFor(g, me.ID)
	if countKind(moves, legal.KindPass) != 1 {
		t.Fatalf("no pass at a table with no requirement: %v", labels(moves))
	}
	for _, m := range moves {
		if m.Kind == legal.KindAttack && m.AlwaysLegal {
			t.Errorf("attack %q marked AlwaysLegal with nothing owed", m.Label)
		}
	}
}
