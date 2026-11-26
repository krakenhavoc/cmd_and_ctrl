package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gatebreaker Ram — Creature — Sheep {2}{G}, 2/2 (Ravnica Allegiance):
//
//	"This creature gets +1/+1 for each Gate you control.
//	 As long as you control two or more Gates, this creature has vigilance
//	 and trample."
//
// Both are statics that re-read the Gates you control on every layer pass.
// The Ram is not a Gate, so it never counts itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b0a97e2-e125-4f91-8581-307e3fa6b048",
		Name:         "Gatebreaker Ram",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+1 for each Gate you control.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := len(effects.PermanentsControlledByMatching(g, source.Controller, effects.HasSubtype("Gate")))
					c.Power += n
					c.Toughness += n
				},
			},
			effects.KeywordGrant(twoOrMoreGates, "vigilance"),
			effects.KeywordGrant(twoOrMoreGates, "trample"),
		},
	})
}

// twoOrMoreGates scopes a grant to this creature while its controller
// controls two or more Gates.
func twoOrMoreGates(target *game.Card, g *game.Game, source *game.Card) bool {
	return effects.SelfOnly(target, g, source) &&
		len(effects.PermanentsControlledByMatching(g, source.Controller, effects.HasSubtype("Gate"))) >= 2
}
