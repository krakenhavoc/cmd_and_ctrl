package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarled Scarhide — Enchantment Creature — Minotaur {B}, 2/1 (Journey into Nyx):
//
//	"Bestow {3}{B}
//	 This creature can't block.
//	 Enchanted creature gets +2/+1 and can't block."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ef73cc0-fe72-4040-a932-bcd4b1a704a3",
		Name:         "Gnarled Scarhide",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{3}{B}"),
		},
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
			effects.PumpAttached(2, 1),
			effects.RestrictAttached(game.CantBlock),
		},
	})
}
