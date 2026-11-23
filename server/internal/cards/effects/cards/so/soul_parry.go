package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Parry — Instant {1}{W}:
//
//	"Prevent all damage one or two target creatures would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): one preventFromSource record per
// targeted creature, each its own source, pinned as the spell resolves
// (CR 400.7). A damage event has one source, so no event meets both
// records and nothing differs from one effect naming both. All damage,
// not only combat damage (its ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b36335b-119b-4f90-a5d7-b77d8f47f55f",
		Name:         "Soul Parry",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve:    effects.ShieldAgainstTargetsSpell(false),
	})
}
