package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadwaters — Sorcery {3}{U} (Magic Origins, S82 older sets):
//
//	"Target player mills X cards, where X is the number of lands you
//	 control."
//
// X is the caster's lands as the spell resolves, whoever the target is. A
// target gone by resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6e70eb3-3da5-4052-9f62-2960a30d4b41",
		Name:         "Dreadwaters",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p := effects.TargetedPlayerOf(ctx)
			if p == uuid.Nil {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, item.Controller, effects.Land()))
			return effects.MillCards{Player: p, N: n}.Apply(ctx)
		},
	})
}
