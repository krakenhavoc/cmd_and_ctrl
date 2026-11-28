package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trumpeting Armodon — Creature — Elephant {3}{G}, 3/3 (Tempest, S82 OLD
// b753, #5628):
//
//	"{1}{G}: Target creature blocks this creature this turn if able."
//
// Lurking Arynx's requirement (TargetBlocksThisCreature) without the
// formidable condition: it waits for the Armodon to attack this turn and
// asks nothing if it never does (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fc3c2a5-2754-4237-8a37-cb3166c20c3c",
		Name:         "Trumpeting Armodon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Target creature blocks this creature this turn if able.",
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetBlocksThisCreature,
		}},
	})
}
