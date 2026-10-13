package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0527b7b5-a55d-4720-af1f-17fc1178f4a6",
		Name:            "Queen Brahne",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Queen Brahne — create a 0/1 Wizard token", func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: NoncreatureCastWizardToken(), N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
