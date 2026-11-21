package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lessons from Life — {2}{G}{U} Sorcery (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Draw three cards. You may put a land card from your hand onto the
//	 battlefield tapped."
//
// Eureka Moment's two steps: the draws first, so a land just drawn is a
// legal pick, then the shared tapped land-from-hand offer, which waits for
// any prompt the draws ask. Putting the land onto the battlefield is not
// playing a land (CR 305.2a), so the turn's land play is still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51df0203-578d-47e1-a9f7-a13445f597d7",
		Name:         "Lessons from Life",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.MayPutALandFromHandTapped("Lessons from Life").Apply(ctx)
			}),
		),
	})
}
