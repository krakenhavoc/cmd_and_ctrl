package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindculling — Sorcery {5}{U} (New Phyrexia, #5575):
//
//	"You draw two cards and target opponent discards two cards."
//
// The two draws happen first, then the target opponent is asked to
// discard two (targetOpponentDiscards), in printed order (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "955a7f4d-7100-4d8b-818c-be8651d0ebe6",
		Name:         "Mindculling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.TargetOpponentDiscards(2)(ctx.Game, item)
		},
	})
}
