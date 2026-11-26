package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Entanglers — Creature — Kor Soldier Ally {4}{W}, 3/4 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control enters, tap
//	 target creature an opponent controls."
//
// One trigger per entering Ally; the target is chosen as the trigger is put
// on the stack, and an illegal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fedd98c-9c58-4ca5-b77b-dcb4cd88d238",
		Name:         "Kor Entanglers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.S82OldB313Rally("Kor Entanglers — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
