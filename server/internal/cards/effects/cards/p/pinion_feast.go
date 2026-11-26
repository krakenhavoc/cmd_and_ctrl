package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinion Feast — Instant {4}{G} (Dragons of Tarkir, S82 old batch 330):
//
//	"Destroy target creature with flying. Bolster 2. (Choose a creature with
//	 the least toughness among creatures you control and put two +1/+1
//	 counters on it.)"
//
// Wing Snare's flying target read through the layered view, destroyed, then
// bolster 2 in printed order (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9812bdf9-98b6-4b4d-8381-a7bbb25069d8",
		Name:         "Pinion Feast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve:    effects.DestroyTargetThenBolster(2),
	})
}
