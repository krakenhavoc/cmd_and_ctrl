package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepfire Elemental — Creature — Elemental {4}{B}{R}, 4/4 (Commander 2013,
// S82 older sets):
//
//	"{X}{X}{1}: Destroy target artifact or creature with mana value X."
//
// Molder's target: the mana value is compared with the X the ability was
// activated with, at announce and again as it resolves (CR 608.2b). A token
// has mana value 0 (CR 202.3), so X = 0 reaches it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8119ebe-aedd-4bdb-8f7f-368674a049fd",
		Name:         "Deepfire Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{1}: Destroy target artifact or creature with mana value X.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{X}{X}{1}"),
			Targets: effects.TargetPermanent("target artifact or creature with mana value X", effects.Or(effects.Artifact(), effects.Creature())).WithManaValueEqualsX(),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
