package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// loyalty_grant_test.go — #4191. The Chain Veil's "{4}, {T}: For each
// planeswalker you control, you may activate one of its loyalty
// abilities once this turn …" buys something only in the bot's own
// main phase, with a planeswalker to use it.

const chainVeilLabel = "{4}, {T}: For each planeswalker you control, you may activate one of its loyalty abilities once this turn as though none of its loyalty abilities have been activated this turn."

func chainVeilView(id string, controller int) protocol.CardView {
	c := rock(id, controller, "The Chain Veil", "{4}")
	c.TypeLine = "Legendary Artifact"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: chainVeilLabel, ManaCost: "{4}", TapCost: true,
		Purpose: &protocol.PurposeView{ExtraLoyaltyActivations: 1},
	}}
	return c
}

// spentWalker is a planeswalker with a +1 row that has used its turn's
// activation.
func spentWalker(id string, controller int) protocol.CardView {
	c := walkerView(id, controller, 4)
	one := 1
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Ref: "own:0", Label: "+1: Draw a card.", LoyaltyCost: &one}}
	c.LoyaltyActivated = true
	return c
}

func chainVeilBoard(withWalker bool, active int, step string) protocol.GameView {
	cards := []protocol.CardView{chainVeilView(cardID(1), 0),
		land(cardID(10), 0), land(cardID(11), 0), land(cardID(12), 0), land(cardID(13), 0)}
	if withWalker {
		cards = append(cards, spentWalker(cardID(2), 0))
	}
	return newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(cards...), withTurn(1, active, step))
}

func TestTheBotActivatesTheChainVeilInItsMainPhaseWithAWalker(t *testing.T) {
	v := chainVeilBoard(true, 0, "precombat_main")
	move := untapMove(t, 0, cardID(1), chainVeilLabel)
	in := input(0, v, passMove(0), move)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != move.Label {
		t.Errorf("the bot chose %q over a second activation for its planeswalker", got)
	}
}

func TestTheBotDoesNotActivateTheChainVeilItCannotUse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		walker bool
		active int
		step   string
	}{
		{"no planeswalker", false, 0, "precombat_main"},
		{"its own combat", true, 0, "declare_attackers"},
		{"an opponent's end step", true, 1, "end"},
	} {
		v := chainVeilBoard(tc.walker, tc.active, tc.step)
		in := input(0, v, passMove(0), untapMove(t, 0, cardID(1), chainVeilLabel))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
			t.Errorf("%s: the bot chose %q; the grant buys nothing there", tc.name, got)
		}
	}
}

// A "0:" loyalty row costs no mana and no tap, but it is not a free
// activation that buys nothing the second time: CR 606.3 bounds it, and
// a second one in a turn is a permission (The Chain Veil, Oath of
// Teferi). The #2777 memory (free_repeat.go) leaves loyalty rows alone,
// because rowCostsOnlyManaAndTaps refuses a loyalty cost; this pins it.
func TestTheBotTakesASecondZeroLoyaltyRowInOneTurn(t *testing.T) {
	const label = "0: Draw a card."
	table := func() aiseat.Input {
		c := walkerView(cardID(2), 0, 4)
		zero := 0
		c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Ref: "own:0", Label: label, LoyaltyCost: &zero}}
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(c), withTurn(4, 0, "precombat_main"))
		return input(0, v, passMove(0), untapMove(t, 0, cardID(2), label))
	}
	p := heuristic.New()
	first := table()
	if got := chose(t, first, decide(t, p, first)); got != label {
		t.Fatalf("first window: the bot chose %q; the control expects it to take the 0 row", got)
	}
	again := table()
	if got := chose(t, again, decide(t, p, again)); got != label {
		t.Errorf("second window: the bot chose %q; a second loyalty activation the engine offers is not a free repeat", got)
	}
}
