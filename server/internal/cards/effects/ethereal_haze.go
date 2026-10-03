package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethereal Haze — Instant — Arcane {W}:
//
//	"Prevent all damage that would be dealt by creatures this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record with no
// source and the property "a creature", checked as the damage would be
// dealt (CR 615.9): combat or not, to anything, for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0cdbd9e7-4941-46ac-99d8-ea227181bdf8",
		Name:         "Ethereal Haze",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldAnything, Queries: []game.PermanentQuery{QueryTypes("creature")}}),
	})
}
