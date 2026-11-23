package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Chang, Cold Warrior — Legendary Creature — Klingon Warrior
// Officer {B}, 1/2:
//
//	"{1}, Sacrifice a creature: Put a +1/+1 counter on each Klingon you
//	 control."
//
// Indulgent Aristocrat's sac outlet for Klingons. The cost may sacrifice
// Chang himself (he is a creature), and then no Klingon he was is left
// to grow. Each Klingon is counted as the ability resolves, through the
// layers, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72217c2e-4f6b-41da-8de1-bb6892d3c452",
		Name:         "General Chang, Cold Warrior",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Put a +1/+1 counter on each Klingon you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect:  effects.TrekB11PutCounterOnEachOfTypeYouControl("Klingon"),
		}},
	})
}
