package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vital Surge — Instant — Arcane {1}{G}:
//
//	"You gain 3 life.
//	Splice onto Arcane {1}{G}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e19c0205-67e3-4e44-995b-751e18fdaafd",
		Name:         "Vital Surge",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{G}"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
