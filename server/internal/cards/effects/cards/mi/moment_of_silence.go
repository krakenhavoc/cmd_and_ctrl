package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moment of Silence — Instant {W} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Target player skips their next combat phase this turn."
//
// A skip record that ends with the turn (S82OldB697TargetPlayerSkipsNextCombatThisTurn):
// cast before the target's combat in their own turn, it removes that combat
// phase; cast in or after it, or in another player's turn, it does nothing
// (CR 500.11, 614.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43af5041-a4bd-4b61-b1a1-bdea916afc59",
		Name:         "Moment of Silence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB697TargetPlayerSkipsNextCombatThisTurn,
	})
}
