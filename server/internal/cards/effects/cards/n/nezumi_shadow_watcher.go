package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nezumi Shadow-Watcher — Creature — Rat Warrior {B}, 1/1 (Betrayers of
// Kamigawa, S82 OLD, #5602):
//
//	"Sacrifice this creature: Destroy target Ninja."
//
// The sacrifice is a cost, so the Shadow-Watcher is gone whatever happens
// to the target; a Ninja that has left does nothing (CR 608.2b). Any
// Ninja, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3eae5419-b686-4321-a89d-cb602c487bfc",
		Name:         "Nezumi Shadow-Watcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target Ninja.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target Ninja", effects.HasSubtype("Ninja")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
