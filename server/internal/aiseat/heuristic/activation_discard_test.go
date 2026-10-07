package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// activation_discard_test.go — #2016: the cards an activated ability's
// discard cost names are priced, not free.

const solphimLabel = "{1}{R/P}{R/P}, Discard two cards: Put an indestructible counter on Solphim, Mayhem Dominus."

func solphim(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := creature(id, controller, "Solphim, Mayhem Dominus", 5, 4, opts...)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Ref: "own:0", Label: solphimLabel}}
	return c
}

func discardActivation(t *testing.T, seat int, src, label string, discards ...string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0, "discard_ids": discards}),
	}
}

func handCreature(id string, seat int, cost string) protocol.CardView {
	c := creature(id, seat, "Hand Creature "+id, 6, 6)
	c.ManaCost = cost
	return c
}

func tappedSolphimBoard(extra ...protocol.CardView) []protocol.CardView {
	return append([]protocol.CardView{solphim(cardID(1), 0, tapped()), land(cardID(10), 0)}, extra...)
}

// activationValue is the value of Solphim's discard activation naming
// these cards, and of passing, in the postcombat main of seat 0's turn.
func activationValue(t *testing.T, hand, bf []protocol.CardView, discards ...string) (act, pass float64) {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)},
		withBattlefield(bf...), withTurn(3, 0, "postcombat_main"))
	in := input(0, v, passMove(0), discardActivation(t, 0, cardID(1), solphimLabel, discards...))
	vals := map[int]float64{}
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1], vals[0]
}

// Discarding two six-drops to pay for the counter costs more than the
// counter is worth; discarding two spare lands late costs almost
// nothing, so the same activation is taken.
func TestActivationDiscardIsPricedAgainstThePayoff(t *testing.T) {
	bf := tappedSolphimBoard(land(cardID(11), 0), land(cardID(12), 0), land(cardID(13), 0),
		land(cardID(14), 0), land(cardID(15), 0), land(cardID(16), 0))
	dear := []protocol.CardView{handCreature(cardID(40), 0, "{4}{R}{R}"), handCreature(cardID(41), 0, "{4}{R}{R}")}
	act, pass := activationValue(t, dear, bf, cardID(40), cardID(41))
	if act >= pass {
		t.Errorf("discarding two 6-drops is valued %.2f, pass %.2f — want it declined", act, pass)
	}
	spare := []protocol.CardView{land(cardID(42), 0), land(cardID(43), 0)}
	act, pass = activationValue(t, spare, bf, cardID(42))
	if act <= pass {
		t.Errorf("discarding one spare land is valued %.2f, pass %.2f — want it taken", act, pass)
	}
}

// The same ability ranks lower the dearer the cards it names.
func TestActivationDiscardCostsMoreForDearerCards(t *testing.T) {
	bf := tappedSolphimBoard()
	hand := []protocol.CardView{handCreature(cardID(40), 0, "{4}{R}{R}"), land(cardID(42), 0), land(cardID(43), 0)}
	dear, _ := activationValue(t, hand, bf, cardID(40), cardID(42))
	cheap, _ := activationValue(t, hand, bf, cardID(42), cardID(43))
	if cheap <= dear {
		t.Errorf("two lands %.2f should outrank a 6-drop and a land %.2f", cheap, dear)
	}
}

// A second indestructible counter on an indestructible Solphim is a
// pure loss, however cheap the discards.
func TestRedundantIndestructibleCounterIsDeclined(t *testing.T) {
	spare := []protocol.CardView{land(cardID(42), 0), land(cardID(43), 0)}
	bf := []protocol.CardView{solphim(cardID(1), 0, tapped(), keywords("indestructible")), land(cardID(10), 0),
		land(cardID(11), 0), land(cardID(12), 0), land(cardID(13), 0), land(cardID(14), 0), land(cardID(15), 0)}
	act, pass := activationValue(t, spare, bf, cardID(42), cardID(43))
	if act >= pass {
		t.Errorf("redundant counter valued %.2f, pass %.2f — want it declined", act, pass)
	}
}

// The #2257 case: after Solphim attacks, with a hand of real cards, the
// bot no longer pays two of them for the counter.
func TestSolphimDoesNotPayRealCardsForTheCounter(t *testing.T) {
	bf := tappedSolphimBoard(land(cardID(11), 0), land(cardID(12), 0), land(cardID(13), 0), land(cardID(14), 0))
	hand := []protocol.CardView{handCreature(cardID(40), 0, "{3}{R}"), handCreature(cardID(41), 0, "{2}{R}")}
	act, pass := activationValue(t, hand, bf, cardID(40), cardID(41))
	if act >= pass {
		t.Errorf("activation %.2f vs pass %.2f — want the pass", act, pass)
	}
}
