package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// reveal_cost_test.go — ADR 0100 amendment 2026-10-07: a reveal or
// behold branch is one announcement that names the card it shows
// (reveal_ids), offered only when the seat has a card to name, and every
// move is one the engine accepts (#544).

const (
	oracleWrensRunVanquisher = "fe9cdd15-a390-4a1e-bae9-474ce8d355b8"
	oracleSilvergillMentor   = "dcf07f38-422f-46c3-aee9-50540b9c9115"
)

type revealPayload struct {
	CostBranch *int     `json:"cost_branch"`
	RevealIDs  []string `json:"reveal_ids"`
}

func revealMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) map[int]revealPayload {
	t.Helper()
	out := map[int]revealPayload{}
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p revealPayload
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

// Wren's Run Vanquisher with an Elf in hand and five Forests: the
// reveal branch names the Elf and the mana branch ({1}{G} plus {3})
// needs all five lands. Neither names a card it should not.
func TestRevealBranchNamesTheCardItShows(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 5, "Forest")
	elf := handCard(seat, game.Card{Name: "Elf Friend", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
	vanquisher := handCard(seat, game.Card{Name: "Wren's Run Vanquisher", TypeLine: "Creature — Elf Warrior", ManaCost: "{1}{G}", OracleID: oracleWrensRunVanquisher})

	moves := legal.EnumerateFor(g, seat.ID)
	got := revealMovesOf(t, moves, vanquisher)
	if len(got) != 2 {
		t.Fatalf("branches offered = %v, want both", got)
	}
	if ids := got[0].RevealIDs; len(ids) != 1 || ids[0] != elf.String() {
		t.Errorf("reveal branch payload = %+v, want the Elf %s (never the Vanquisher itself)", got[0], elf)
	}
	if len(got[1].RevealIDs) != 0 {
		t.Errorf("mana branch payload = %+v, want no reveal", got[1])
	}
	dispatchAll(t, g, seat.ID, moves)
}

// With no other Elf the reveal branch is not offered; with two lands
// the mana branch is not either, so the seat is offered no Vanquisher.
func TestRevealBranchNeedsACardToShow(t *testing.T) {
	for _, withElf := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		basicLands(g, seat, 2, "Forest")
		if withElf {
			handCard(seat, game.Card{Name: "Elf Friend", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
		}
		vanquisher := handCard(seat, game.Card{Name: "Wren's Run Vanquisher", TypeLine: "Creature — Elf Warrior", ManaCost: "{1}{G}", OracleID: oracleWrensRunVanquisher})
		moves := legal.EnumerateFor(g, seat.ID)
		got := revealMovesOf(t, moves, vanquisher)
		if withElf {
			if _, ok := got[0]; !ok || len(got) != 1 {
				t.Fatalf("with an Elf: branches = %v, want only the reveal", got)
			}
		} else if len(got) != 0 {
			t.Fatalf("with no Elf and two lands: branches = %v, want none", got)
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}

// Behold prefers a permanent already on the table to a card in hand: it
// shows nothing new.
func TestBeholdPrefersAPermanent(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Island")
	handCard(seat, game.Card{Name: "Merfolk Pal", TypeLine: "Creature — Merfolk", ManaCost: "{U}"})
	onTable := battlefieldCard(g, seat, game.Card{Name: "Merfolk Scout", TypeLine: "Creature — Merfolk", ManaCost: "{U}", Power: 1, Toughness: 1})
	mentor := handCard(seat, game.Card{Name: "Silvergill Mentor", TypeLine: "Creature — Merfolk Wizard", ManaCost: "{1}{U}", OracleID: oracleSilvergillMentor})

	moves := legal.EnumerateFor(g, seat.ID)
	got := revealMovesOf(t, moves, mentor)
	if ids := got[0].RevealIDs; len(ids) != 1 || ids[0] != onTable.String() {
		t.Fatalf("behold payload = %+v, want the permanent %s", got[0], onTable)
	}
	dispatchAll(t, g, seat.ID, moves)
}
