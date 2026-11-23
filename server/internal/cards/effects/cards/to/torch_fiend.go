package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch Fiend — Creature — Devil {1}{R}, 2/1 (Dark Ascension, #5571):
//
//	"{R}, Sacrifice this creature: Destroy target artifact."
//
// Goblin Replica's activation for {R}: the Fiend is sacrificed as the cost
// (CR 602.2b), and an artifact gone by resolution is left alone (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15f8049c-f21d-4cd0-aa81-1a533936cf77",
		Name:         "Torch Fiend",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target artifact.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
