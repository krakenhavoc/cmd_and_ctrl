package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wiccan, Rising Magician — {4}{U} Legendary Creature — Mutant Warlock
// Hero 4/4 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Whenever you cast a noncreature spell, exile another nonland,
//	 nontoken permanent. Return that card to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// Eerie Interlude's exile-then-delayed-return (exileTargetsThenScheduleReturn)
// behind a cast trigger that targets: the permanent is picked as the trigger
// goes on the stack and re-checked as it resolves (CR 608.2b), so one that
// left in response stays where it is. The target is any controller's
// permanent, never Wiccan himself (Another) and never a land or a token.
// The card returns under its owner's control as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc17b7a3-7711-4449-80bd-692e8ed9087d",
		Name:            "Wiccan, Rising Magician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Noncreature(), "Wiccan, Rising Magician — exile another nonland, nontoken permanent; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Wiccan, Rising Magician — return the exiled permanent")
					}),
				effects.Another(effects.TargetPermanent("another target nonland, nontoken permanent", effects.Nonland(), effects.Not(effects.IsTokenPredicate())))),
		},
	})
}
