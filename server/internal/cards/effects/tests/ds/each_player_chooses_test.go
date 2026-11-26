package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const eachPlayerChoosesTestOracle = "test-each-player-chooses-war-or-peace"

// TestEachPlayerChoosesAsEntersRecordsEveryAnswerForTheStatics is the
// helper behind Archangel of Strife (#6097): as the permanent enters,
// every player chooses war or peace in APNAP order (CR 101.4), each answer
// is recorded on the permanent for that player (game.Card.PlayerNotes),
// and a static reads it: creatures controlled by players who chose war get
// +3/+0, those of players who chose peace +0/+3.
func TestEachPlayerChoosesAsEntersRecordsEveryAnswerForTheStatics(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: eachPlayerChoosesTestOracle,
		Name:     "War or Peace Test",
		AsEnters: effects.EachPlayerChoosesAsEnters("War or Peace Test", "war", "peace"),
		Static: []game.StaticAbility{
			{
				Label: "Creatures controlled by players who chose war get +3/+0.", Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(c *game.Card, _ *game.Game, source *game.Card) bool {
					return c.IsCreature() && source.PlayerNoted(c.Controller, "war")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power += 3 },
			},
			{
				Label: "Creatures controlled by players who chose peace get +0/+3.", Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(c *game.Card, _ *game.Game, source *game.Card) bool {
					return c.IsCreature() && source.PlayerNoted(c.Controller, "peace")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Toughness += 3 },
			},
		},
	})
	g := effectstest.NewCatalogGame(t)
	active := g.Turn.ActiveSeat
	var order []*game.Player
	for i := range g.Seats {
		order = append(order, g.Seats[(active+i)%len(g.Seats)])
	}
	bears := map[uuid.UUID]uuid.UUID{}
	for _, p := range order {
		bears[p.ID] = effectstest.B12Creature(g, p.ID, "Bear", "Creature — Bear", 2, 2)
	}
	angel := effectstest.RfdCast(t, g, "War or Peace Test", "Creature — Angel", eachPlayerChoosesTestOracle, 5, 5)

	// APNAP: the active player first, then the others in turn order; war
	// for the first and third, peace for the others.
	for i, p := range order {
		for _, other := range order {
			if other != p && effectstest.LatestOptionPickFor(g, other.ID) != nil {
				t.Fatalf("player %d is asked while player %d's answer is owed", indexOf(order, other), i)
			}
		}
		effectstest.AnswerOptionPick(t, g, p.ID, i%2)
	}
	for i, p := range order {
		v := effectstest.LayeredCard(t, g, bears[p.ID])
		wantP, wantT := 2, 5
		if i%2 == 0 {
			wantP, wantT = 5, 2
		}
		if v.CurrentPower() != wantP || v.CurrentToughness() != wantT {
			t.Errorf("player %d's Bear is %d/%d, want %d/%d", i, v.CurrentPower(), v.CurrentToughness(), wantP, wantT)
		}
	}
	if v := effectstest.LayeredCard(t, g, angel); v.CurrentPower() != 8 || v.CurrentToughness() != 5 {
		t.Errorf("the active player's own creature is %d/%d, want 8/5 (they chose war)", v.CurrentPower(), v.CurrentToughness())
	}
}

func indexOf(ps []*game.Player, p *game.Player) int {
	for i, q := range ps {
		if q == p {
			return i
		}
	}
	return -1
}
