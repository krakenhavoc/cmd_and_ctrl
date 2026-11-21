package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indestructible Aura — Instant {W}:
//
//	"Prevent all damage that would be dealt to target creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the recipient family of the
// not-one-use shield. A preventFromSource record pinned to the target
// with no source: every instance of damage to it this turn, from
// anything, is prevented. A creature that leaves the battlefield and
// returns is a new object the shield no longer covers (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e10e8d56-bba6-412d-970e-c24969f32b5b",
		Name:         "Indestructible Aura",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShieldTargetSpell(),
	})
}
