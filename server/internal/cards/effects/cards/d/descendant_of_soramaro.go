package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Descendant of Soramaro — Creature — Human Wizard {3}{U}, 2/3 (Saviors of
// Kamigawa, #5601):
//
//	"{1}{U}: Look at the top X cards of your library, where X is the
//	 number of cards in your hand, then put them back in any order."
//
// X is counted as the ability resolves (CR 608.2h); the look and the
// reorder are Portent's (LookAtLibraryThenPlace). An empty hand, or an
// empty library, looks at nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cf3126e-fbac-4245-beba-3d4eedda24bf",
		Name:         "Descendant of Soramaro",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Look at the top X cards of your library, where X is the number of cards in your hand, then put them back in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.SokB1CardsInYourHand(ctx)
				if x <= 0 {
					return nil
				}
				return effects.LookAtLibraryThenPlace{
					N:         x,
					Placement: game.LibraryPlaceTop,
					Label:     "Descendant of Soramaro — put them back in any order",
				}.Apply(ctx)
			},
		}},
	})
}
