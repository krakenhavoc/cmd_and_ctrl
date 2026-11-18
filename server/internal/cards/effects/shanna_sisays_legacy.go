package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shanna, Sisay's Legacy — {G}{W} Legendary Creature — Human Warrior, 0/0:
//
//	"Shanna can't be the target of abilities your opponents control.
//	 Shanna gets +1/+1 for each creature you control."
//
// The targeting clause refuses abilities only, not spells, and only an
// opponent's (rprB03AbilitiesFromOpponentsCantTargetThis). The boost is
// counted on every layer pass (7c), Shanna herself included, so she is
// always at least 1/1 while on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "546201d0-3fa5-44e4-a526-ab0945711e49",
		Name:         "Shanna, Sisay's Legacy",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Shanna gets +1/+1 for each creature you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
				c.Power += n
				c.Toughness += n
			},
		}},
		TargetingRestrictions: []game.TargetingRestriction{
			rprB03AbilitiesFromOpponentsCantTargetThis("Shanna can't be the target of abilities your opponents control."),
		},
	})
}
