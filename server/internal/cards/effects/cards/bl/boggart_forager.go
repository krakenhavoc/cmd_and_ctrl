package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Forager — Creature — Goblin Rogue {R}, 1/1 (Lorwyn, S82 OLD b547,
// #5593):
//
//	"{R}, Sacrifice this creature: Target player shuffles their library."
//
// Soldier of Fortune's shuffle with the sacrifice as part of the cost: the
// ability still resolves once the Forager is gone, and a target player who
// has left the game by resolution makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cef9e0d3-5e28-4c5f-88fc-0d78af5a4db1",
		Name:         "Boggart Forager",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Target player shuffles their library.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB547TargetPlayerShufflesTheirLibrary,
		}},
	})
}
