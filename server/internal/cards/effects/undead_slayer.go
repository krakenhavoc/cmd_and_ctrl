package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Slayer — Creature — Human Cleric {2}{W}, 2/2 (Magic 2010, #5585):
//
//	"{W}, {T}: Exile target Skeleton, Vampire, or Zombie."
//
// Phoenix Down's exile bullet as a tap ability: the target must be a
// Skeleton, Vampire or Zombie on activation and again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08fc069e-7fdd-4726-aa08-a96bbd2bcaa3",
		Name:         "Undead Slayer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Exile target Skeleton, Vampire, or Zombie.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetPermanent("target Skeleton, Vampire, or Zombie", AnySubtype("Skeleton", "Vampire", "Zombie")),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
