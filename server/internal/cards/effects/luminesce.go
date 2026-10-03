package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luminesce — Instant {W}:
//
//	"Prevent all damage that black sources and red sources would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record with no
// source and a colour property, checked as the damage would be dealt (CR
// 615.9): any source that is black or red then, permanent or spell, to
// anything, for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7b0f482d-35a3-47f1-8283-0fd47e30cc31",
		Name:         "Luminesce",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldAnything, Queries: []game.PermanentQuery{QueryColors("B", "R")}}),
	})
}
