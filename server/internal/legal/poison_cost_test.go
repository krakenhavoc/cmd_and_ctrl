package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

var poisonCostThen = game.RegisterConfirmThen("test/4613-poison-cost-legal", func(*game.Game, game.ConfirmAnswer) error { return nil })

// #4613: the accept move of a confirm that gives poison counters carries the
// count as MoveCost.Poison, at any poison total (a player can always get
// counters), and the decline carries no price.
func TestConfirmAcceptMoveCarriesThePoisonItGives(t *testing.T) {
	for _, have := range []int{0, 5, 9} {
		g := newTable(t)
		payer := g.Seats[1]
		if have > 0 {
			if err := g.SetPoison(payer.ID, have); err != nil {
				t.Fatal(err)
			}
		}
		g.WithWriteLock(func() {
			if _, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
				Chooser:      payer.ID,
				Question:     "Ward — get 5 poison counters, or it is countered",
				AcceptLabel:  "Get 5 poison counters",
				DeclineLabel: "Let it be countered",
				PoisonCost:   5,
			}, poisonCostThen, nil); err != nil {
				t.Fatal(err)
			}
		})
		moves := legal.EnumerateFor(g, payer.ID)
		accepts, declines := 0, 0
		for _, m := range moves {
			var p struct {
				Apply *bool `json:"apply"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			switch {
			case p.Apply != nil && *p.Apply:
				accepts++
				if m.Cost == nil || m.Cost.Poison != 5 || m.Cost.Life != 0 {
					t.Errorf("at %d poison the accept move's cost is %+v, want Poison 5", have, m.Cost)
				}
			case p.Apply != nil:
				declines++
				if m.Cost != nil {
					t.Errorf("the decline carries a price: %+v", m.Cost)
				}
			}
		}
		if accepts != 1 || declines != 1 {
			t.Errorf("at %d poison: %d accept and %d decline moves, want one each", have, accepts, declines)
		}
		dispatchAll(t, g, payer.ID, moves)
	}
}
