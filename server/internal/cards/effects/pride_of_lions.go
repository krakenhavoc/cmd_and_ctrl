package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pride of Lions — Creature — Cat {3}{G}{G}, 4/4 (Starter 1999, #5620):
//
//	"You may have this creature assign its combat damage as though it
//	 weren't blocked."
//
// Wolverine, Claws Out's static: a layer 6 ability on the Pride sets
// Characteristic.AssignsAsThoughUnblocked on the Pride alone, an option
// over the blocked assignment of CR 510.1c. For a blocked Pride the
// combat damage step asks its controller whether to divide its damage
// among its blockers as usual or put all of it on the player,
// planeswalker or battle it is attacking. A Pride that has lost its
// abilities loses the option with them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b5e0df8-f8f0-40a1-9814-89d73a026ad0",
		Name:         "Pride of Lions",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have this creature assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
	})
}
