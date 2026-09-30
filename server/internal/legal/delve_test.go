package legal_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// delve_test.go — ADR 0100 §6 and owner decision 3: the enumerator
// offers a delve cast (CR 702.66a) TWO payments — the fewest graveyard
// cards that make it affordable, and the full budget — and the engine
// accepts both.

const oracleTreasureCruise = "5b6bdf5a-2742-4851-92cd-a857a3852836"

type delvePayload struct {
	DelveIDs []string `json:"delve_ids"`
}

// delveBoard is Treasure Cruise ({7}{U}) in hand, `islands` Islands to
// pay with and `fuel` cards in the graveyard.
func delveBoard(t *testing.T, islands, fuel int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, islands, "Island")
	for i := 0; i < fuel; i++ {
		graveyardCard(seat, game.Card{Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}"})
	}
	cruise := handCard(seat, game.Card{
		Name: "Treasure Cruise", TypeLine: "Sorcery", ManaCost: "{7}{U}", OracleID: oracleTreasureCruise,
	})
	return g, seat, cruise
}

func delvePaymentsOf(t *testing.T, moves []legal.Move, src uuid.UUID) [][]string {
	t.Helper()
	var out [][]string
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p delvePayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, p.DelveIDs)
	}
	return out
}

// Two Islands and six graveyard cards: the Islands pay {U} and one
// generic, so the fewest payment IS the full graveyard, and the two
// payments are one move.
func TestDelveOffersOnePaymentWhenFewestIsFull(t *testing.T) {
	g, seat, cruise := delveBoard(t, 2, 6)
	moves := legal.EnumerateFor(g, seat.ID)
	pays := delvePaymentsOf(t, moves, cruise)
	if len(pays) != 1 || len(pays[0]) != 6 {
		t.Fatalf("delve payments = %v, want one move exiling all six", pays)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Three Islands and seven graveyard cards: the fewest payment exiles
// five (the lands pay {U} and two generic), the full one exiles seven,
// and both are offered and both are accepted by the engine (#544).
func TestDelveOffersFewestAndFullPayments(t *testing.T) {
	g, seat, cruise := delveBoard(t, 3, 7)
	moves := legal.EnumerateFor(g, seat.ID)
	pays := delvePaymentsOf(t, moves, cruise)
	if len(pays) != 2 {
		t.Fatalf("delve payments = %v, want two (fewest and full)", pays)
	}
	if len(pays[0]) != 5 || len(pays[1]) != 7 {
		t.Fatalf("delve payment sizes = %d and %d, want 5 then 7", len(pays[0]), len(pays[1]))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// One Island and three graveyard cards cannot pay {7}{U} however much
// is delved, so no cast is offered.
func TestDelveOffersNothingItCannotAfford(t *testing.T) {
	g, seat, cruise := delveBoard(t, 1, 3)
	if pays := delvePaymentsOf(t, legal.EnumerateFor(g, seat.ID), cruise); len(pays) != 0 {
		t.Fatalf("delve payments = %v, want none", pays)
	}
}

// The pool arrives in the seat's fuel order: with a policy that prices
// one card as precious, the fewest payment leaves it in the graveyard.
func TestDelvePaymentTakesTheCheapestFuelFirst(t *testing.T) {
	g, seat, cruise := delveBoard(t, 3, 6)
	precious := seat.Graveyard.Cards[0].InstanceID
	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{
		OrderCostFuel: func(c legal.TargetCandidate) float64 {
			if c.ID == precious {
				return 100
			}
			return 1
		},
	})
	pays := delvePaymentsOf(t, moves, cruise)
	if len(pays) == 0 || len(pays[0]) != 5 {
		t.Fatalf("delve payments = %v, want the fewest to exile five", pays)
	}
	for _, id := range pays[0] {
		if id == precious.String() {
			t.Fatal("the fewest payment ate the card the policy prices highest")
		}
	}
}
