package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigil of the Nayan Gods — Enchantment — Aura {1}{G}{W} (Alara Reborn):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each creature you control.
//	 Cycling {G/W} ({G/W}, Discard this card: Draw a card.)"
//
// The count is the Aura controller's creatures on every layer pass
// (Crown of Gondor's PumpAttachedPer), the enchanted creature included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92017087-18f4-46ee-8808-5959d31d0011",
		Name:         "Sigil of the Nayan Gods",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
			}),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{G/W}")},
	})
}
