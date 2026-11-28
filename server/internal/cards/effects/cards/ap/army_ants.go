package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Army Ants — Creature — Insect {1}{B}{R}, 1/1 (Visions, S82 older sets batch
// 790, #5631):
//
//	"{T}, Sacrifice a land: Destroy target land."
//
// The land sacrificed is a cost, chosen as the ability is activated, and may be
// the very land that is then targeted only if it is still there to target:
// a sacrificed land is gone before targets are legal (CR 602.2b). The ability
// needs the Ants untapped and free of summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c39112ab-ee1f-4f00-b18b-692b5fe32b80",
		Name:         "Army Ants",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a land: Destroy target land.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.B08SacrificeALand()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
