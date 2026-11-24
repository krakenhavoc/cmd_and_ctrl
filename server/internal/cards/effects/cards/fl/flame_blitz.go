package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Blitz — Enchantment {R} (Modern Horizons 2, #5501):
//
//	"At the beginning of your end step, this enchantment deals 5 damage to
//	 each planeswalker.
//	 Cycling {2}"
//
// Every planeswalker takes the damage, yours included, as one damage event.
// If Flame Blitz has left the battlefield by resolution it deals the damage
// as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "817eb91a-e3b4-40c3-9a6e-e4d0387e0f4f",
		Name:         "Flame Blitz",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Flame Blitz — 5 damage to each planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.Planeswalker(), 5)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
