package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arlinn, Voice of the Pack — Legendary Planeswalker {4}{G}{G}, loyalty 7
// (Jumpstart 2022, #5948):
//
//	"Each creature you control that's a Wolf or a Werewolf enters with an
//	 additional +1/+1 counter on it.
//	 −2: Create a 2/2 green Wolf creature token."
//
// Dragonstorm Globe's CR 614 entry replacement, with the match widened to
// two creature types (read through the source's text, CR 612.2). The
// Wolf token it makes enters under the same replacement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a15b83df-91b7-49f0-9bec-5776a3dac8bd",
		Name:            "Arlinn, Voice of the Pack",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 7,
		Replacements: []game.ReplacementEffect{
			effects.B28PermanentsYouControlEnterWithACounter("Arlinn, Voice of the Pack: a Wolf or Werewolf creature enters with an additional +1/+1 counter",
				func(c game.Card) bool {
					return c.IsCreature() && (c.HasSubtypeWord("Wolf") || c.HasSubtypeWord("Werewolf"))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Create a 2/2 green Wolf creature token.",
			Cost:    effects.LoyaltyCost(-2),
			Purpose: game.Purpose{Tokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}),
		}},
	})
}
