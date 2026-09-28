package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// optional_cost_clause_test.go — #1716: a target clause widened by an
// announced optional cost. Too Evil to Stay Dead's teamwork lifts the
// "mana value 4 or less" ceiling, so the bot must be offered the big
// creature card ONLY on a teamwork cast, and every move it is offered
// must dispatch (#544's agreement).

const oracleTooEvilToStayDead = "e50fecff-8872-42f5-8882-41ad13d9d1ae"

func TestTooEvilToStayDeadOffersTheBigCardOnlyWithTeamwork(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Too Evil to Stay Dead", TypeLine: "Sorcery", OracleID: oracleTooEvilToStayDead, ManaCost: "{2}{B}",
	})
	lands(g, active, "Swamp", "Swamp", 3)
	battlefieldCard(g, active, creature("Ogre", "{2}{R}", 4, 4))
	small := graveyardCard(active, creature("Small Bear", "{1}{G}", 2, 2))
	big := graveyardCard(active, creature("Big Beater", "{3}{G}{G}", 5, 5))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
	var paidBig, paidSmall, unpaidSmall int
	for _, m := range moves {
		var p struct {
			OptionalCosts []int `json:"optional_costs"`
			Targets       []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %q: %v", m.Label, err)
		}
		if len(p.Targets) != 1 {
			t.Errorf("%q: want exactly one target, got %v", m.Label, p.Targets)
			continue
		}
		paid := len(p.OptionalCosts) > 0
		switch p.Targets[0].ID {
		case big.String():
			if !paid {
				t.Errorf("%q: an unpaid cast must never target the mana-value-5 card", m.Label)
			}
			paidBig++
		case small.String():
			if paid {
				paidSmall++
			} else {
				unpaidSmall++
			}
		}
	}
	if paidBig == 0 || paidSmall == 0 || unpaidSmall == 0 {
		t.Errorf("want teamwork casts at both cards and a plain cast at the small one; got big %d, small %d / %d; labels: %v",
			paidBig, paidSmall, unpaidSmall, labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
