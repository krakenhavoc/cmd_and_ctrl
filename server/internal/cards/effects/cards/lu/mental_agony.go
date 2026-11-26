package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mental Agony — {3}{B} Sorcery: "Target player discards two cards and loses
// 2 life."
//
// The discard is the player's own choice (Mind Rot's body), then the life
// loss, which is life loss and not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e80a7118-dd6b-4907-ad69-09387e0c2114",
		Name:         "Mental Agony",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2, LifeLoss: 2}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.MindRotOnResolve(item, ctx); err != nil {
				return err
			}
			return effects.DrainTargetedOpponent(2)(ctx.Game, item)
		},
	})
}
