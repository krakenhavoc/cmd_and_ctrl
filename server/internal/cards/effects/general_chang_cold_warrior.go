package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "72217c2e-4f6b-41da-8de1-bb6892d3c452",
		Name:         "General Chang, Cold Warrior",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Put a +1/+1 counter on each Klingon you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect:  trekB11PutCounterOnEachOfTypeYouControl("Klingon"),
		}},
	})
}
