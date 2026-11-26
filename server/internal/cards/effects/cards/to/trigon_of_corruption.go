package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trigon of Corruption — Artifact {4} (Scars of Mirrodin, S82 OLD batch 470,
// #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {B}{B}, {T}: Put a charge counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Put a -1/-1 counter
//	 on target creature."
//
// The Trigon shape (OldB470Trigon): three counters as it enters, refilled for
// its colour, one removed as the cost of the effect (Nightmare Sower's counter). A target that has left
// by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.OldB470Trigon(
		effects.Spec{OracleID: "0da19cd4-7b3f-47bb-bb35-cbcc88ab53cd", Name: "Trigon of Corruption", Completeness: effects.CompletenessFull},
		"{B}{B}", "Put a -1/-1 counter on target creature.",
		effects.TargetCreature("target creature"), game.Purpose{Answers: game.AnswerRemove},
		effects.R35NightmareSowerCounter))
}
