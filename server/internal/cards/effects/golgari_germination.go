package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3f9cf871-7df6-49d5-aab3-ad945571b707",
		Name:         "Golgari Germination",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead)
			}, "Golgari Germination — create a 1/1 green Saproling", Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})),
		},
	})
}
