package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Spiral — Instant {4}{U} (Return to Ravnica, #5566):
//
//	"Shuffle all cards from your graveyard into your library. Target player
//	 mills that many cards."
//
// Learn from the Past's shuffle of your own graveyard, counting the cards as
// they move; the target then mills that many (CR 701.17a). The shuffle happens
// whether or not the target is still legal; the mill needs a legal target
// (CR 608.2b). The spell is on the stack, not in the graveyard, so it is not
// counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f928e54-3027-4f2f-9fef-d63b9ca416aa",
		Name:         "Psychic Spiral",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 0
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
				n = len(p.Graveyard.Cards)
			}
			if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, ctx.Controller()); err != nil {
				return err
			}
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer && ctx.IsTargetLegal(t) {
				return effects.MillCards{Player: t.ID, N: n}.Apply(ctx)
			}
			return nil
		},
	})
}
