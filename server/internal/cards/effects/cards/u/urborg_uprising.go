package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Uprising — Sorcery {4}{B} (Dominaria Remastered, #5948):
//
//	"Return up to two target creature cards from your graveyard to your hand.
//	 Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "882d3be7-1c0f-4d2b-8f2e-32369488ef82",
		Name:         "Urborg Uprising",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnLegalGraveyardTargetsToHand(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
