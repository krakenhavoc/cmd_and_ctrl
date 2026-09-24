package protocol

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blockers_view_test.go pins the #328 wire contract: the snapshot has
// to tell the client which seats owe a declare-blockers decision,
// because the client's auto-pass cannot work it out for itself
// (blocking is a turn-based action, not a response) and must not
// re-derive block legality in TypeScript.

// stageBlockingWindow puts `attacker`'s creature into combat against
// `defender`, who gets one untapped creature, and parks the cursor on
// declare_blockers. Mirrors the board in the bug report.
func stageBlockingWindow(t *testing.T, g *game.Game, attackerSeat, defenderSeat int) uuid.UUID {
	t.Helper()
	attacker := g.Seats[attackerSeat]
	defender := g.Seats[defenderSeat]

	atk := game.NewCard("Enduring Curiosity", attacker.ID)
	atk.TypeLine = "Creature — Cat"
	atk.Power, atk.Toughness = 4, 4
	g.Battlefield.PushTop(atk)

	blk := game.NewCard("Aang, Swift Savior", defender.ID)
	blk.TypeLine = "Legendary Creature — Human"
	blk.Power, blk.Toughness = 3, 3
	g.Battlefield.PushTop(blk)

	g.Turn.ActiveSeat = attackerSeat
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(atk.InstanceID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into declare_blockers: %v", err)
	}
	return blk.InstanceID
}

func TestViewCarriesBlockDecisionSeats(t *testing.T) {
	g := buildActiveGame(t)
	blockerID := stageBlockingWindow(t, g, 1, 0)

	v := ViewOfGame(g)
	if v.Turn.Step != "declare_blockers" {
		t.Fatalf("step: got %q, want declare_blockers", v.Turn.Step)
	}
	if got := v.Turn.BlockDecisionSeats; len(got) != 1 || got[0] != 0 {
		t.Fatalf("block_decision_seats: got %v, want [0]", got)
	}

	// It must survive the per-viewer projection for BOTH viewers —
	// attackers and untapped creatures are public information, and the
	// defender's client is the one that reads it.
	for _, seat := range g.Seats {
		fv := ViewOfGameFor(g, seat.ID.String())
		if got := fv.Turn.BlockDecisionSeats; len(got) != 1 || got[0] != 0 {
			t.Errorf("viewer %s: block_decision_seats: got %v, want [0]", seat.Name, got)
		}
	}

	// And it has to actually reach the wire under the name the client
	// reads.
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"block_decision_seats":[0]`)) {
		t.Errorf("block_decision_seats missing from JSON")
	}

	// Tapping the defender's only creature removes the legal block, so
	// the field drops out entirely (omitempty) and auto-pass is free
	// to skip a window that costs the player nothing.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == blockerID {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	after := ViewOfGame(g)
	if got := after.Turn.BlockDecisionSeats; len(got) != 0 {
		t.Errorf("tapped-only defender: got %v, want none", got)
	}
	rawAfter, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(rawAfter, []byte("block_decision_seats")) {
		t.Errorf("empty block_decision_seats must be omitted from the wire")
	}
}

func TestViewOmitsBlockDecisionSeatsOutsideTheStep(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	if got := v.Turn.BlockDecisionSeats; len(got) != 0 {
		t.Errorf("outside declare_blockers: got %v, want none", got)
	}
}

// #1279: the turn cursor says where each defender's block declaration
// stands — pending until they finish, then declared — and a finished
// defender leaves block_decision_seats even with a creature at home.
func TestViewCarriesBlockDeclarationStatus(t *testing.T) {
	g := buildActiveGame(t)
	stageBlockingWindow(t, g, 1, 0)

	v := ViewOfGame(g)
	if got := v.Turn.BlockPendingSeats; len(got) != 1 || got[0] != 0 {
		t.Fatalf("block_pending_seats = %v, want [0]", got)
	}
	if len(v.Turn.BlocksDeclaredSeats) != 0 {
		t.Fatalf("blocks_declared_seats = %v, want none yet", v.Turn.BlocksDeclaredSeats)
	}

	if err := g.FinishBlocks(g.Seats[0].ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	v = ViewOfGame(g)
	if len(v.Turn.BlockPendingSeats) != 0 {
		t.Errorf("block_pending_seats = %v, want none after finishing", v.Turn.BlockPendingSeats)
	}
	if got := v.Turn.BlocksDeclaredSeats; len(got) != 1 || got[0] != 0 {
		t.Errorf("blocks_declared_seats = %v, want [0]", got)
	}
	if len(v.Turn.BlockDecisionSeats) != 0 {
		t.Errorf("block_decision_seats = %v: a defender who has declared owes nothing", v.Turn.BlockDecisionSeats)
	}
	raw, err := json.Marshal(FilterViewFor(v, g.Seats[1].ID.String()))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"blocks_declared_seats":[0]`)) {
		t.Errorf("the status is public: it must survive the per-viewer filter; got %s", raw)
	}
}
