package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eerie Interference — Instant {2}{W}:
//
//	"Prevent all damage that would be dealt to you and creatures you control this turn by creatures."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record with no
// source and the property "a creature", checked as the damage would be
// dealt (CR 615.9), protecting you and the creatures you control as the
// damage would be dealt, so a creature that comes under your control
// later this turn is protected too.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "35ba0e70-da7c-4ed9-b8ec-7dd5c5ce110e",
		Name:         "Eerie Interference",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldYouAndYourCreatures, Queries: []game.PermanentQuery{QueryTypes("creature")}}),
	})
}
