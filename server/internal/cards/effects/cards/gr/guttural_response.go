package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guttural Response — Instant {R/G} (Shadowmoor, #5591):
//
//	"Counter target blue instant spell."
//
// Gainsay's blue-spell target clause narrowed to instants: the target is
// legal only while the spell is blue and an instant (CR 608.2b re-checks it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fb80575-9c8e-4745-9a8f-d48cdcbbb868",
		Name:         "Guttural Response",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target blue instant spell", effects.And(effects.OfColor("U"), effects.Instant())),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
