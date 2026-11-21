package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Queen Brahne — Legendary Creature — Human Noble {2}{R}, 2/1:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature
//	 gets +1/+1 until end of turn.)
//	 Whenever Queen Brahne attacks, create a 0/1 black Wizard creature
//	 token with "Whenever you cast a noncreature spell, this token deals
//	 1 damage to each opponent.""
//
// Prowess is the printed keyword; the Wizard is the shared
// NoncreatureCastWizardToken, whose ping lives on the token itself
// (ADR 0083). The token is not created tapped or attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0527b7b5-a55d-4720-af1f-17fc1178f4a6",
		Name:            "Queen Brahne",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Queen Brahne — create a 0/1 Wizard token", func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.NoncreatureCastWizardToken(), N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
