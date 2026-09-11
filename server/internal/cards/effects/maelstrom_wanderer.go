package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maelstrom Wanderer — Legendary Creature — Elemental {5}{G}{U}{R},
// 7/5:
//
//	"Creatures you control have haste.
//	 Cascade, cascade"
//
// The sprint's exit criterion in one card, and the reason
// TriggeredAbility is a SLICE: "cascade, cascade" is two separate
// triggered abilities, not one that runs twice. Both go on the stack
// when the Wanderer is cast, the controller orders them (CR 603.3b),
// and each exiles until it finds something costing less than eight —
// which is very nearly anything.
//
// The haste grant is an ordinary Layer 6 static. "Creatures you
// control" includes the Wanderer itself, which matters on the turn it
// lands and on the two free spells that come with it.
func init() {
	Register(Spec{
		OracleID: "ad9b7fbc-61c8-43ee-a65c-99206fd1e4df",
		Name:     "Maelstrom Wanderer",
		Triggered: []game.TriggeredAbility{
			Cascade(),
			Cascade(),
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "haste" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "haste")
			},
		}},
	})
}
