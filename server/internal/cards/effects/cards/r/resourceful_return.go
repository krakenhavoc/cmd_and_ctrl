package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resourceful Return — Sorcery {1}{B} (Aether Revolt, S82 older sets, #5536):
//
//	"Return target creature card from your graveyard to your hand. If you
//	 control an artifact, draw a card."
//
// The artifact is looked for as the spell resolves (CR 608.2c), after the
// creature card has been returned. A target that has left the graveyard
// fizzles the whole spell (CR 608.2b), so there is no draw then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "952ef791-5c89-4965-9f34-0228acb18bc4",
		Name:         "Resourceful Return",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnLegalGraveyardTargetsToHand(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.CountControlled(ctx.Game, ctx.Controller(), effects.MatchArtifact) == 0 {
					return nil
				}
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
