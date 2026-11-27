package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Three Tragedies — Sorcery — Arcane {3}{B}{B} (Betrayers of Kamigawa,
// #5602):
//
//	"Target player discards three cards."
//
// Mindstab's and Fugue's three-card discard: the targeted player picks the
// cards (CR 701.9b), and a hand of fewer than three is discarded whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b30e80ef-a4d7-48e2-918e-a347c2d9da59",
		Name:         "Three Tragedies",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerDiscards(3)(ctx.Game, item)
		},
	})
}
