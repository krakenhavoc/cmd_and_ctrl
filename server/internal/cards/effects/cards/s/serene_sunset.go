package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serene Sunset — Instant {X}{G}:
//
//	"Prevent all combat damage X target creatures would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the target count is the announced
// X (CountFromX). Each target gets a combat-damage shield with itself as
// the source, pinned as the spell resolves (CR 400.7) — one record per
// source, which nothing a player can see tells from one effect, since an
// event has one source.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2ef027d-0ef7-4c8b-a594-3918290b02d9",
		Name:         "Serene Sunset",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
		OnResolve:    effects.ShieldAgainstTargetsSpell(true),
	})
}
