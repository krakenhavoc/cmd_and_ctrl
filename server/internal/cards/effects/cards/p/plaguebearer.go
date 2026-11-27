package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plaguebearer — Creature — Zombie {1}{B}, 1/1 (Exodus, S82 OLD batch 744, #5626):
//
//	"{X}{X}{B}: Destroy target nonblack creature with mana value X."
//
// Two X slots, Dauntless Dismantler's: X=2 costs {4}{B}. The clause admits only
// a nonblack creature whose mana value is exactly the announced X, judged as the
// ability is activated and again as it resolves (CR 608.2b). The destruction can
// be regenerated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8cc1dc4-740d-4c5b-ac51-7355730a6595",
		Name:         "Plaguebearer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{B}: Destroy target nonblack creature with mana value X.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{X}{X}{B}"),
			Targets: effects.TargetCreature("target nonblack creature with mana value X", effects.NonBlack()).WithManaValueEqualsX(),
			Effect:  effects.S82OldB744DestroyTheTargetCreature,
		}},
	})
}
