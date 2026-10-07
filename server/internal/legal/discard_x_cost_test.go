package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_x_cost_test.go — #2527: "Discard X cards" announces its count
// as X (CR 602.2b), so the enumerator offers a bounded ladder of
// payments — the cheapest 1, 2, 3 cards — each with x_value equal to
// the cards named, never the X=0 no-op for a card whose whole effect
// is X (#810), and nothing at all when the hand is empty. Every move is
// one the engine accepts (#544).

const oracleGixYawgmothPraetor = "928d977e-cff0-4e0e-83bb-16d73a754f35"

type discardXPayload struct {
	XValue     int      `json:"x_value"`
	DiscardIDs []string `json:"discard_ids"`
}

func gixDiscardMoves(t *testing.T, handSize int) (*game.Game, *game.Player, []legal.Move, []discardXPayload) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 7, "Swamp")
	gix := battlefieldCard(g, seat, game.Card{
		Name: "Gix, Yawgmoth Praetor", TypeLine: "Legendary Creature — Phyrexian Praetor",
		ManaCost: "{1}{B}{B}", OracleID: oracleGixYawgmothPraetor, Power: 3, Toughness: 3,
	})
	for i := 0; i < handSize; i++ {
		handCard(seat, game.Card{Name: "Fodder", TypeLine: "Sorcery", ManaCost: "{2}"})
	}
	moves := legal.EnumerateFor(g, seat.ID)
	var payloads []discardXPayload
	for _, m := range moves {
		if m.Kind != legal.KindActivate || m.Source != gix {
			continue
		}
		var p discardXPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		payloads = append(payloads, p)
	}
	return g, seat, moves, payloads
}

// With five cards in hand the ladder is 1, 2, 3: never 0 (the no-op),
// never past the cap, x_value always the cards named.
func TestDiscardXOffersALadderOfCounts(t *testing.T) {
	g, seat, moves, got := gixDiscardMoves(t, 5)
	counts := map[int]bool{}
	for _, p := range got {
		if p.XValue != len(p.DiscardIDs) {
			t.Errorf("x_value %d does not match the %d cards named", p.XValue, len(p.DiscardIDs))
		}
		counts[len(p.DiscardIDs)] = true
	}
	for _, n := range []int{1, 2, 3} {
		if !counts[n] {
			t.Errorf("count %d not offered (offered %v)", n, counts)
		}
	}
	if counts[0] || counts[4] || counts[5] {
		t.Errorf("counts offered outside the ladder: %v", counts)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A hand of one offers the one count it can pay.
func TestDiscardXWithAShortHand(t *testing.T) {
	_, _, _, got := gixDiscardMoves(t, 1)
	if len(got) == 0 {
		t.Fatal("no activation offered with one card in hand")
	}
	for _, p := range got {
		if len(p.DiscardIDs) != 1 || p.XValue != 1 {
			t.Errorf("payload %+v, want exactly the one-card payment", p)
		}
	}
}

// An empty hand pays only X=0, which does nothing, so nothing is offered.
func TestDiscardXWithAnEmptyHandIsNotOffered(t *testing.T) {
	_, _, _, got := gixDiscardMoves(t, 0)
	if len(got) != 0 {
		t.Errorf("offered %d activations with an empty hand: %+v", len(got), got)
	}
}
