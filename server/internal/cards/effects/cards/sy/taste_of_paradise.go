package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taste of Paradise — Sorcery {3}{G} (Alliances, S82 OLD batch 820):
//
//	"As an additional cost to cast this spell, you may pay {1}{G} any number
//	 of times.
//	 You gain 3 life plus an additional 3 life for each additional {1}{G}
//	 you paid."
//
// That is a multikicker by another name (CR 702.33d): the engine caps one
// announcement at twenty payments, as it does for every multikicker.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c9a20d03-b90a-43df-8ab4-ff0b38c90a08",
		Name:          "Taste of Paradise",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{1}{G}", 20)},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: 3 + 3*ctx.KickedTimes()}.Apply(ctx)
		},
	})
}
