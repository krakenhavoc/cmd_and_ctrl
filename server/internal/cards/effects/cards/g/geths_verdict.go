package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geth's Verdict — Instant {B}{B}:
//
//	"Target player sacrifices a creature of their choice and loses 1 life."
//
// Both halves are the one sentence: the life is lost whether or not a
// creature was sacrificed.
func init() {
	const name = "Geth's Verdict"
	effects.Register(effects.Spec{
		OracleID:     "0f575f18-4606-4b85-9a90-07f8d8d46d06",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetPlayerSacrificesACreature(name+" — sacrifice a creature")(ctx.Game, item); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -1)
				}
			}
			return nil
		},
	})
}
