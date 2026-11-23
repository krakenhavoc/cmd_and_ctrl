package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redeem — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt this turn to up to two target
//	 creatures."
//
// ADR 0108 §7, Delivery PR 7 (#1904): one prevention effect, so one
// not-one-use record protecting every chosen creature
// (ShieldTheTargetPermanents). A target gone by resolution is not
// protected, and the others still are (CR 608.2b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2032f88-a000-4ed2-a7e5-61d29723d45e",
		Name:         "Redeem",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve:    effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldTheTargetPermanents}),
	})
}
