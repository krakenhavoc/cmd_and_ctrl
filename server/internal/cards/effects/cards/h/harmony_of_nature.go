package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harmony of Nature — Sorcery {2}{G} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Tap any number of untapped creatures you control. You gain 4 life for each creature tapped this way."
//
// Jaddi Lifestrider's shape: a choose-any-number prompt, then 4 life for each creature that really became tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c943c9d-090b-41c1-9003-24c91601e589",
		Name:         "Harmony of Nature",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB735TapAnyNumberForLife(ctx.Game, item)
		},
	})
}
