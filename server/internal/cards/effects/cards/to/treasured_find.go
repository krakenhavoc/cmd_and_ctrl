package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasured Find — Sorcery {B}{G} (Return to Ravnica, #5566):
//
//	"Return target card from your graveyard to your hand. Exile Treasured
//	 Find."
//
// Once and Future's two steps: the card returns if it is still a legal target
// (CR 608.2b), then the spell exiles itself instead of going to the graveyard.
// A spell that is countered or fizzles never resolves, so it is not exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e99d6c5c-b029-4cd1-bd7a-6f8ea91eb4a7",
		Name:         "Treasured Find",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		Targets:      effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnTargetFromGraveyardToHand(ctx.Item, ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
