package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Symbiotic Deployment — Enchantment {2}{G} (Apocalypse, S82 OLD, #5614):
//
//	"Skip your draw step.
//	 {1}, Tap two untapped creatures you control: Draw a card."
//
// Necropotence's skipped draw step as a replacement while the Enchantment is
// out (CR 614.10), and an ability paid by tapping two creatures, which is not
// {T} so a summoning-sick creature may pay it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6872d7a-b647-4baf-bd1e-06707ccbb71a",
		Name:         "Symbiotic Deployment",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SkipYourDrawStep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Tap two untapped creatures you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.OldB629TapTwoUntappedCreaturesCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
