package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mouser Mark III — {1}{U/R} Artifact Creature — Robot 2/3 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"This creature can't attack unless you control another artifact."
//
// Chief Warg's Company's shape: a layer 6 static on the creature itself
// that sets CantAttack while its controller controls no OTHER artifact (it
// is an artifact itself and does not count). The other permanents' types
// are read after the layers, as Doctor Doom's check reads them, so an
// artifact made by an effect counts. Restrictions are checked as attackers
// are declared (CR 508.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad0c6d46-a1ca-4a9f-93f0-4a4dc31868e6",
		Name:         "Mouser Mark III",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't attack unless you control another artifact.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !selfOnly(target, g, source) {
					return false
				}
				for _, id := range permanentsControlledByMatching(g, source.Controller, Artifact()) {
					if id != source.InstanceID {
						return false
					}
				}
				return true
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
	})
}
