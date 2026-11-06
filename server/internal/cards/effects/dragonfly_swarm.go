package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonfly Swarm — Creature — Dragon Insect {1}{U}{R}, */3 (Avatar: The
// Last Airbender, #4459):
//
//	"Flying, ward {1}
//	 This creature's power is equal to the number of noncreature, nonland
//	 cards in your graveyard.
//	 When this creature dies, if there's a Lesson card in your graveyard,
//	 draw a card."
//
// The power is a layer 7a characteristic-defining ability (CR 604.3) over
// its controller's graveyard, recounted whenever a graveyard changes. The
// dies trigger is Walltop Sentries' intervening "if" (CR 603.4) with a
// draw in place of the gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83abf1d0-04e4-49e5-bf63-59f77829ffd1",
		Name:            "Dragonfly Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature's power is equal to the number of noncreature, nonland cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = dragonflySwarmNoncreatureNonlandCards(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Dragonfly Swarm — ward {1}"),
			avB16WhenThisDiesIfLesson("Dragonfly Swarm — draw a card", DrawCards{N: 1}),
		},
	})
}

// dragonflySwarmNoncreatureNonlandCards counts the noncreature, nonland
// cards in a player's graveyard.
func dragonflySwarmNoncreatureNonlandCards(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() && !c.IsLand() {
			n++
		}
	}
	return n
}
