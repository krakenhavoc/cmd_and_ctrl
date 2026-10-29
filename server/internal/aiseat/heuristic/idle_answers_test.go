package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestAnswerOnlyRowsWaitForSomethingToAnswer(t *testing.T) {
	for _, answer := range []string{"survive", "protect", "prevent", "restrict"} {
		t.Run(answer, func(t *testing.T) {
			for _, sacrifice := range []bool{true, false} {
				c := savior("creatures", false, answer)
				c.Tapped = true
				c.ActivatedAbilities[0].SacrificeSelf = sacrifice
				v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
					withBattlefield(append(manaLands(3, 0, 100), c, ogre(cardID(2)), ogre(cardID(3)))...), withTurn(5, 0, "postcombat_main"))
				m := activate(t, 0, c.InstanceID, "Answer", nil)
				in := input(0, v, passMove(0), m)
				if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
					t.Errorf("%s (sacrifice=%v): chose %q with nothing to answer", answer, sacrifice, got)
				}
				cfg := heuristic.DefaultConfig()
				cfg.PriceAnswerOnlyRows = false
				if got := chose(t, in, decide(t, heuristic.NewWithConfig(cfg), in)); got != m.Label {
					t.Errorf("back-out chose %q, want the idle activation", got)
				}
				if got := rankValue(t, heuristic.NewWithConfig(cfg), in, m.Label); got <= 0 {
					t.Errorf("back-out should retain a positive activation, got %g", got)
				}
				if got := rankValue(t, heuristic.New(), in, m.Label); got >= 0 {
					t.Errorf("%s (sacrifice=%v) scores %g, want below pass", answer, sacrifice, got)
				}
			}
		})
	}
}

func TestAnswerOnlySacrificePaysForItsSourceInResponse(t *testing.T) {
	c := savior("creatures", false, "restrict")
	div, item := divination(1, cardID(9))
	v := respond([]protocol.CardView{c}, div, item)
	m := activate(t, 0, c.InstanceID, "Restrict", nil)
	in := input(0, v, passMove(0), m)
	if got := rankValue(t, heuristic.New(), in, m.Label); got >= 0 {
		t.Errorf("unpriced restriction sacrifices a creature for %g, want a loss", got)
	}
}

// A declared payoff keeps its price, even when the row also answers.
func TestIdleAnswerGuardPreservesRowsWithValue(t *testing.T) {
	c := savior("creatures", false, "survive")
	c.ActivatedAbilities[0].SacrificeSelf = false
	c.ActivatedAbilities[0].Purpose.Draws = 2
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(c), withTurn(5, 0, "postcombat_main"))
	m := activate(t, 0, c.InstanceID, "Draw and protect", nil)
	in := input(0, v, passMove(0), m)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != m.Label {
		t.Errorf("chose %q over a useful draw", got)
	}
}
