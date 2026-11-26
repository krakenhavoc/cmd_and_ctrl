package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raised by Wolves — {3}{G}{G} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 When this Aura enters, create two 2/2 green Wolf creature tokens.
//	 Enchanted creature gets +1/+1 for each Wolf you control."
//
// The count is the Aura's controller's Wolves (CR 109.5), read through the layers on
// every pass: the two tokens, and any other Wolf, the enchanted creature itself
// included if it is one.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9e364e6-6202-4859-9f65-69c9951959ac",
		Name:         "Raised by Wolves",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 2, CreatureTokens: 2},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool {
					return c.IsCreature() && c.HasSubtypeWord("Wolf")
				})
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Raised by Wolves — create two 2/2 green Wolf creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 2})),
		},
	})
}
