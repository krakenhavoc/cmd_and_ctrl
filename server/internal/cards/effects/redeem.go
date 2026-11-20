package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c2032f88-a000-4ed2-a7e5-61d29723d45e",
		Name:         "Redeem",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldTheTargetPermanents}),
	})
}
