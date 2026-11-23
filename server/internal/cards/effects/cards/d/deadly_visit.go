package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadly Visit — Sorcery {3}{B}{B} (Guilds of Ravnica, #5525):
//
//	"Destroy target creature.
//	 Surveil 2."
//
// The shared destroyTargetThenSurveil: the surveil happens whether or not
// the creature was destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04a3bb42-0464-4638-a6ac-d5f412c222cc",
		Name:         "Deadly Visit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTargetThenSurveil(2),
	})
}
