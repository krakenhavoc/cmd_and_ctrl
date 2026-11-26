package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiss of the Amesha — Sorcery {4}{W}{U} (Battlebond, #5529):
//
//	Target player gains 7 life and draws two cards.
//
// The target is re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "779d3e58-de06-44be-98f8-f05f7711ecb2",
		Name:         "Kiss of the Amesha",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who := effects.FirstLegalTargetPlayer(ctx)
			if who == uuid.Nil {
				return nil
			}
			if err := (effects.GainLife{Player: who, Amount: 7}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: who, N: 2}.Apply(ctx)
		},
	})
}
