package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// either_cost_test.go — ADR 0100 §6: an either/or additional cost is
// one announcement per branch the seat can pay, each priced with its
// own cost_branch through the one pricer, and every move is one the
// engine accepts (#544).

const (
	oracleLightningAxe  = "81b90905-fbc0-426a-a084-c3300533abb4"
	oracleDemandAnswers = "c11e84a1-dbda-429b-8cd6-fd0deaefc689"
	oracleBitterTriumph = "776341cb-d2ec-423f-9250-92dc8bd8d503"
)

type branchPayload struct {
	CostBranch   *int     `json:"cost_branch"`
	DiscardIDs   []string `json:"discard_ids"`
	SacrificeIDs []string `json:"sacrifice_ids"`
}

func branchMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) map[int]branchPayload {
	t.Helper()
	out := map[int]branchPayload{}
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p branchPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		if p.CostBranch == nil {
			t.Fatalf("move %q carries no cost_branch", m.Label)
		}
		out[*p.CostBranch] = p
	}
	return out
}

// Lightning Axe with six Mountains and a card to discard: both branches
// are offered, the discard branch names a card, and the mana branch —
// priced {5}{R} — needs all six lands.
func TestEitherCostOffersEachPayableBranch(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 6, "Mountain")
	handCard(seat, game.Card{Name: "Pitch", TypeLine: "Instant", ManaCost: "{1}"})
	battlefieldCard(g, g.Seats[1], creature("Ogre", "{2}", 2, 2))
	axe := handCard(seat, game.Card{Name: "Lightning Axe", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningAxe})

	moves := legal.EnumerateFor(g, seat.ID)
	got := branchMovesOf(t, moves, axe)
	if len(got) != 2 {
		t.Fatalf("branches offered = %v, want both", got)
	}
	if len(got[0].DiscardIDs) != 1 {
		t.Errorf("discard branch payload = %+v, want one discard", got[0])
	}
	if len(got[1].DiscardIDs) != 0 {
		t.Errorf("mana branch payload = %+v, want no discard", got[1])
	}
	dispatchAll(t, g, seat.ID, moves)
}

// With one Mountain the mana branch is unaffordable and is not offered;
// with no other card in hand the discard branch is unpayable and is not
// offered either — so a seat with one Mountain and an empty hand is
// offered no Lightning Axe at all (CR 601.2h).
func TestEitherCostOffersNoUnpayableBranch(t *testing.T) {
	for _, withPitch := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		basicLands(g, seat, 1, "Mountain")
		if withPitch {
			handCard(seat, game.Card{Name: "Pitch", TypeLine: "Instant", ManaCost: "{1}"})
		}
		battlefieldCard(g, g.Seats[1], creature("Ogre", "{2}", 2, 2))
		axe := handCard(seat, game.Card{Name: "Lightning Axe", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningAxe})
		moves := legal.EnumerateFor(g, seat.ID)
		got := branchMovesOf(t, moves, axe)
		if withPitch {
			if _, ok := got[0]; !ok || len(got) != 1 {
				t.Fatalf("with a card to discard: branches = %v, want only the discard", got)
			}
		} else if len(got) != 0 {
			t.Fatalf("with nothing to discard and one land: branches = %v, want none", got)
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}

// Demand Answers' sacrifice branch pays with an artifact, never a land
// or a creature, and its discard branch with a card in hand.
func TestEitherCostPaysTheSacrificeBranch(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Mountain")
	relic := battlefieldCard(g, seat, game.Card{Name: "Relic", TypeLine: "Artifact"})
	battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2))
	handCard(seat, game.Card{Name: "Pitch", TypeLine: "Instant", ManaCost: "{1}"})
	da := handCard(seat, game.Card{Name: "Demand Answers", TypeLine: "Instant", ManaCost: "{1}{R}", OracleID: oracleDemandAnswers})
	moves := legal.EnumerateFor(g, seat.ID)
	got := branchMovesOf(t, moves, da)
	if p, ok := got[0]; !ok || len(p.SacrificeIDs) != 1 || p.SacrificeIDs[0] != relic.String() {
		t.Fatalf("sacrifice branch = %+v, want the artifact", got[0])
	}
	if p, ok := got[1]; !ok || len(p.DiscardIDs) != 1 {
		t.Fatalf("discard branch = %+v, want one discard", got[1])
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The bot never pays life down to zero: Bitter Triumph's "pay 3 life"
// is not offered to a seat at 3 life, though CR 119.4 would allow it.
// The move that pays life says so in its cost.
func TestEitherCostLifeBranchPolicy(t *testing.T) {
	for _, life := range []int{3, 4} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		seat.Life = life
		basicLands(g, seat, 2, "Swamp")
		battlefieldCard(g, g.Seats[1], creature("Ogre", "{2}", 2, 2))
		bt := handCard(seat, game.Card{Name: "Bitter Triumph", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: oracleBitterTriumph})
		moves := legal.EnumerateFor(g, seat.ID)
		got := branchMovesOf(t, moves, bt)
		_, offered := got[1]
		if offered != (life > 3) {
			t.Fatalf("at %d life the life branch offered = %v", life, offered)
		}
		for _, m := range moves {
			if m.Source == bt && m.Kind == legal.KindCast && (m.Cost == nil || m.Cost.Life != 3) {
				t.Errorf("move %q cost = %+v, want Life 3", m.Label, m.Cost)
			}
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}
