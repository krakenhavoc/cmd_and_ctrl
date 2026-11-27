package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Offering — {4}{B}{B} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Destroy target nonblack creature. You gain 3 life."
//
// Solemn Offering's destroy-then-gain (DestroyTargetThenGainLife) on a
// nonblack creature; a target no longer legal makes the spell do nothing, the
// life included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "102356a9-0de5-4034-b121-034da5deb4f2",
		Name:         "Dark Offering",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
