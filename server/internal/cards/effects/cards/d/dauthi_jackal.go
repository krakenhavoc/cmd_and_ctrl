package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Jackal — Creature — Dauthi Jackal {2}{B}, 2/1 (Exodus, S82 OLD batch 744, #5626):
//
//	"Shadow
//	 {B}{B}, Sacrifice this creature: Destroy target blocking creature."
//
// The sacrifice is a cost, so the Jackal is gone before the ability resolves;
// the target must be a creature blocking at that moment, re-checked as the
// ability resolves (CR 608.2b). The destruction can be regenerated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8847b057-8896-4c72-a11f-3c183b1cc60e",
		Name:            "Dauthi Jackal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, Sacrifice this creature: Destroy target blocking creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect:  effects.S82OldB744DestroyTheTargetCreature,
		}},
	})
}
