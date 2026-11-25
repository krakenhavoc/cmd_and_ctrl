package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angrath's Ambusher — Creature — Orc Pirate {2}{B}, 2/3:
//
//	"This creature gets +2/+0 as long as you control an Angrath planeswalker."
//
// A layer-7c pump on the creature itself while its controller controls a
// planeswalker with the Angrath subtype, read live (CR 613.4c), so it comes
// and goes with the walker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85791f11-44ab-4405-b284-1d91c7e0afbb",
		Name:         "Angrath's Ambusher",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +2/+0 as long as you control an Angrath planeswalker.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfCreature(target, g, source) && effects.ControlsA("Angrath")(g, source.Controller)
			},
			Apply: effects.PumpBy(2, 0),
		}},
	})
}
