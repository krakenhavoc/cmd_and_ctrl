package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Condescend — Instant {X}{U} (Jumpstart 2022, #5948):
//
//	"Counter target spell unless its controller pays {X}. Scry 2."
//
// Clash of Wills' pay-unless with X announced at the cast, then the caster's
// scry. With the spell's one target illegal as it resolves the whole spell
// does nothing (CR 608.2b), so there is no scry either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96cf3c10-733d-4110-9dab-43d0cd4e6629",
		Name:         "Condescend",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetSpell("target spell"),
		Purpose:      game.Purpose{Scry: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.GapB172CondescendBody(item, ctx, ctx.X())
		},
	})
}
