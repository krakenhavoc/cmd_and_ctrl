package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// land_sacrifice_test.go — #2469. Harrow ("sacrifice a land; put up to
// two basic lands onto the battlefield untapped") was cast in 30% of the
// games it was offered in, against A3's 50% bar. Its sacrifice is on the
// non-mana cost list, so ADR 0126 §5's leftover windows never lowered
// its bar: outside a main phase it needed InstantThreshold (1.50), which
// two lands for one land and a card does not reach once the ramp premium
// has half closed.

func harrow(id string, controller int) protocol.CardView {
	c := spell(id, controller, "Harrow", "{2}{G}")
	c.Purpose = &protocol.PurposeView{Lands: 2}
	return c
}

func sacrificeCast(t *testing.T, seat int, card protocol.CardView, sacrificed ...string) legal.Move {
	id := card.InstanceID
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast, Label: "Cast " + card.Name,
		Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "sacrifice_ids": sacrificed}),
	}
}

// The bot holds a five-drop and has four lands (the first of them the
// one it sacrifices), in the end step before its turn.
func harrowTable(t *testing.T, card protocol.CardView, extra ...protocol.CardView) (aiseat.Input, string) {
	lands := manaLands(4, 2, 100)
	bf := append(lands, extra...)
	in := input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(card, rampFiveDrop(cardID(9), 2))), newSeat(3)},
		withBattlefield(bf...), withTurn(5, 1, "end")),
		passMove(2), sacrificeCast(t, 2, card, lands[0].InstanceID))
	return in, "Cast " + card.Name
}

func TestTheBotCastsHarrowInTheEndStepBeforeItsTurn(t *testing.T) {
	in, label := harrowTable(t, harrow(cardID(1), 2))
	v := rankValue(t, heuristic.New(), in, label)
	if v <= 0 || v > 1.5 {
		t.Fatalf("Harrow priced %+.2f; the test needs a price above passing and below InstantThreshold", v)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != label {
		t.Errorf("chose %q (Harrow priced %+.2f), want Harrow: it ends with a land more than it began with, untapped (#2469)", got, v)
	}
	// The baseline is the pre-S66 heuristic and keeps the old bar.
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), in)); got != "Pass priority" {
		t.Errorf("baseline chose %q, want the pass", got)
	}
}

// A sacrifice that its own declared lands do not more than replace keeps
// the normal bar: one land for one land, or no lands declared at all.
func TestALandSacrificeThatIsNotNetManaKeepsTheNormalBar(t *testing.T) {
	for name, p := range map[string]*protocol.PurposeView{
		"one for one":    {Lands: 1, Draws: 1},
		"no lands":       {Draws: 2},
		"nothing at all": nil,
	} {
		c := harrow(cardID(1), 2)
		c.Name, c.Purpose = "Look-alike", p
		in, label := harrowTable(t, c)
		v := rankValue(t, heuristic.New(), in, label)
		if name != "nothing at all" && (v <= 0 || v > 1.5) {
			t.Fatalf("%s: priced %+.2f; the test needs a price above passing and below InstantThreshold", name, v)
		}
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("%s: chose %q at %+.2f, want the pass", name, got, v)
		}
	}
}

// Sacrificing something that is not a land keeps the normal bar, even
// for a card that declares lands.
func TestAHarrowThatSacrificesACreatureKeepsTheNormalBar(t *testing.T) {
	c := harrow(cardID(1), 2)
	c.Purpose = &protocol.PurposeView{Lands: 2}
	imp := creature(cardID(30), 2, "Imp", 0, 1)
	lands := manaLands(4, 2, 100)
	in := input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(c, rampFiveDrop(cardID(9), 2))), newSeat(3)},
		withBattlefield(append(lands, imp)...), withTurn(5, 1, "end")),
		passMove(2), sacrificeCast(t, 2, c, imp.InstanceID))
	v := rankValue(t, heuristic.New(), in, "Cast Harrow")
	if v <= 0 || v > 1.5 {
		t.Fatalf("priced %+.2f; the test needs a price above passing and below InstantThreshold", v)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("chose %q at %+.2f: a sacrificed creature is not a land the cast replaces", got, v)
	}
}
