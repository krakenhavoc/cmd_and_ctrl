package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// teamwork_modes_test.go — #1703 with #1655: HULK SMASH!'s teamwork
// cast is offered with BOTH bullets and a team, its unpaid cast with
// exactly one, and every move dispatches.

const oracleHulkSmash = "5f97d49c-d0fa-4776-9455-93c1a172cd83"

func TestHulkSmashTeamworkCastTakesBothBullets(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	hulk := handCard(active, game.Card{
		Name: "HULK SMASH!", TypeLine: "Instant", OracleID: oracleHulkSmash, ManaCost: "{1}{R}",
	})
	lands(g, active, "Mountain", "Mountain", 2)
	battlefieldCard(g, active, creature("Ogre", "{2}{R}", 4, 4))
	battlefieldCard(g, opp, creature("Elf", "{G}", 1, 1))
	battlefieldCard(g, opp, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), hulk)
	paid, unpaid := 0, 0
	for _, m := range moves {
		var p struct {
			Modes         []int    `json:"modes"`
			OptionalCosts []int    `json:"optional_costs"`
			TeamworkIDs   []string `json:"teamwork_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %q: %v", m.Label, err)
		}
		if len(p.OptionalCosts) > 0 {
			paid++
			if len(p.Modes) != 2 || len(p.TeamworkIDs) != 1 {
				t.Errorf("%q: a teamwork cast must take both bullets with the Ogre: modes %v, team %v", m.Label, p.Modes, p.TeamworkIDs)
			}
			continue
		}
		unpaid++
		if len(p.Modes) != 1 {
			t.Errorf("%q: an unpaid cast takes exactly one bullet, got %v", m.Label, p.Modes)
		}
	}
	if paid == 0 || unpaid == 0 {
		t.Errorf("want both paid and unpaid casts, got %d / %d; labels: %v", paid, unpaid, labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
