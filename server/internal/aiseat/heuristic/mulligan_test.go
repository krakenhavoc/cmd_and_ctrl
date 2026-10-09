package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2693: the mulligan counted lands only, so review game 2 kept a
// two-land seven whose cheapest spells cost three and missed its next
// two land drops.

func mulliganMoves(t *testing.T, next int) (legal.Move, legal.Move) {
	keep := legal.Move{Type: legal.TypeKeepHand, Player: seatID(0), Kind: legal.KindMulligan, Label: "Keep hand"}
	mull := legal.Move{
		Type: legal.TypeMulligan, Player: seatID(0), Kind: legal.KindMulligan, Label: "Mulligan",
		Params: mustJSON(t, map[string]int{"hand_size": next}),
	}
	return keep, mull
}

func withCost(cost string) cardOpt { return func(c *protocol.CardView) { c.ManaCost = cost } }

// game2Hand is seq 8's hand with Exotic Orchard as a second Mountain
// and the spells' costs as printed.
func game2Hand(extra ...protocol.CardView) []protocol.CardView {
	hand := []protocol.CardView{
		land(cardID(1), 0),
		land(cardID(2), 0),
		creature(cardID(3), 0, "Angrath's Marauders", 6, 6, withCost("{5}{R}{R}")),
		creature(cardID(4), 0, "Goldspan Dragon", 4, 4, withCost("{3}{R}{R}")),
		creature(cardID(5), 0, "Solphim, Mayhem Dominus", 5, 5, withCost("{2}{R}{R}")),
		spell(cardID(6), 0, "Chaos Warp", "{2}{R}"),
		spell(cardID(7), 0, "Unexpected Windfall", "{2}{R}{R}"),
	}
	return append(hand, extra...)
}

func decideMulligan(t *testing.T, p *heuristic.Policy, seat protocol.PlayerView, next int) string {
	t.Helper()
	keep, mull := mulliganMoves(t, next)
	in := input(0, newView([]protocol.PlayerView{seat, newSeat(1)}), keep, mull)
	return chose(t, in, decide(t, p, in))
}

func TestMulliganTakesTheFreeOneOnAHandThatCannotCastSoon(t *testing.T) {
	seat := newSeat(0, withHand(game2Hand()...))
	if got := decideMulligan(t, heuristic.New(), seat, 7); got != "Mulligan" {
		t.Fatalf("chose %q on two lands and nothing under three mana; the free mulligan is better", got)
	}
	if got := decideMulligan(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), seat, 7); got != "Keep hand" {
		t.Fatalf("baseline chose %q; with KeepNeedsCast off the hand is kept on its land count", got)
	}
}

func TestMulliganKeepsTheFloorHandWithATwoDrop(t *testing.T) {
	// A Signet replaces Windfall: castable off the two lands in hand.
	hand := game2Hand()
	hand[6] = spell(cardID(7), 0, "Arcane Signet", "{2}")
	if got := decideMulligan(t, heuristic.New(), newSeat(0, withHand(hand...)), 7); got != "Keep hand" {
		t.Fatalf("chose %q on two lands and a two-drop", got)
	}
}

func TestMulliganCountsColoursAtTheFloor(t *testing.T) {
	// Two Mountains and a Counterspell: two mana, but no blue source.
	hand := game2Hand()
	hand[6] = spell(cardID(7), 0, "Counterspell", "{U}{U}")
	if got := decideMulligan(t, heuristic.New(), newSeat(0, withHand(hand...)), 7); got != "Mulligan" {
		t.Fatalf("chose %q on a blue two-drop with only red sources", got)
	}
	// A Mountain and an Island and a {1}{U} spell: castable.
	hand[1] = land(cardID(2), 0, func(c *protocol.CardView) {
		c.Name, c.TypeLine = "Island", "Basic Land — Island"
		c.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: "{U}"}}
	})
	hand[6] = spell(cardID(7), 0, "Impulse", "{1}{U}")
	if got := decideMulligan(t, heuristic.New(), newSeat(0, withHand(hand...)), 7); got != "Keep hand" {
		t.Fatalf("chose %q on a {1}{U} spell with an Island in hand", got)
	}
}

func TestMulliganReadsAFetchLandAsAnyColour(t *testing.T) {
	// A land with no mana ability on the view (a fetch) can find the
	// colour; erring toward a keep is the old behaviour.
	hand := game2Hand()
	hand[1] = land(cardID(2), 0, func(c *protocol.CardView) {
		c.Name, c.TypeLine, c.ManaAbilities = "Polluted Delta", "Land", nil
	})
	hand[6] = spell(cardID(7), 0, "Mana Leak", "{1}{U}")
	if got := decideMulligan(t, heuristic.New(), newSeat(0, withHand(hand...)), 7); got != "Keep hand" {
		t.Fatalf("chose %q on a fetch land and a {1}{U} spell", got)
	}
}

func TestMulliganKeepsTheFloorHandWhenTheMulliganCostsACard(t *testing.T) {
	// After the free mulligan the next one draws six: the check does not
	// apply, and the land count keeps the hand as before.
	seat := newSeat(0, withHand(game2Hand()...))
	seat.MulligansTaken = 1
	if got := decideMulligan(t, heuristic.New(), seat, 6); got != "Keep hand" {
		t.Fatalf("chose %q; a mulligan to six is not spent on a hand with its lands", got)
	}
}

func TestMulliganCheckIsOnlyAtTheLandFloor(t *testing.T) {
	// Three lands and the same top-heavy spells: kept on lands alone.
	hand := game2Hand()
	hand[6] = land(cardID(7), 0)
	if got := decideMulligan(t, heuristic.New(), newSeat(0, withHand(hand...)), 7); got != "Keep hand" {
		t.Fatalf("chose %q on three lands", got)
	}
}

func TestMulliganReachWidensWhatCountsAsSoon(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	cfg.KeepCastReach = 1
	// Chaos Warp, three mana, is within two lands plus one.
	if got := decideMulligan(t, heuristic.NewWithConfig(cfg), newSeat(0, withHand(game2Hand()...)), 7); got != "Keep hand" {
		t.Fatalf("chose %q with a reach of one and a three-drop in hand", got)
	}
}
