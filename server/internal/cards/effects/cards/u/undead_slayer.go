package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Slayer — Creature — Human Cleric {2}{W}, 2/2 (Magic 2010, #5585):
//
//	"{W}, {T}: Exile target Skeleton, Vampire, or Zombie."
//
// Phoenix Down's exile bullet as a tap ability: the target must be a
// Skeleton, Vampire or Zombie on activation and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08fc069e-7fdd-4726-aa08-a96bbd2bcaa3",
		Name:         "Undead Slayer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Exile target Skeleton, Vampire, or Zombie.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Skeleton, Vampire, or Zombie", effects.AnySubtype("Skeleton", "Vampire", "Zombie")),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
