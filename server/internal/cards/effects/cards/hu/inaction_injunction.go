package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inaction Injunction — Sorcery {1}{U} (Return to Ravnica, #5566):
//
//	"Detain target creature an opponent controls.
//	 Draw a card."
//
// One target: when it is illegal on resolution the whole spell does nothing
// (CR 608.2b), so no card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f6bb2dd-c888-43bd-b9ef-8bd451ab669b",
		Name:         "Inaction Injunction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := effects.OldB423DetainTargets("Inaction Injunction — detain")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
