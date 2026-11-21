package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Germination — Enchantment {1}{B}{G} (Ravnica: City of Guilds,
// #5600):
//
//	"Whenever a nontoken creature you control dies, create a 1/1 green
//	 Saproling creature token."
//
// Blight Mound's death watch: a nontoken creature that dies under its
// controller's control makes one Saproling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f9cf871-7df6-49d5-aab3-ad945571b707",
		Name:         "Golgari Germination",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.IsToken(dead)
			}, "Golgari Germination — create a 1/1 green Saproling", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
		},
	})
}
