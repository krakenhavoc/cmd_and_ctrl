package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Violent Eruption — Instant {1}{R}{R}{R} (#1658, unblocked by #1656):
//
//	"Violent Eruption deals 4 damage divided as you choose among any
//	 number of targets.
//	 Madness {1}{R}{R}"
//
// Fury's fixed-4 "any number" division (bounded above by the 4 points,
// since every target needs at least 1) on an instant instead of a
// creature's ETB, plus the madness alternative cost — both halves of
// which are existing machinery (Spec.Madness; #657's madness engine).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "842578ca-0a86-4e96-bfd5-45931488f7c1",
		Name:         "Violent Eruption",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{R}{R}",
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
