package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omen of the Forge — Enchantment {1}{R} (Theros Beyond Death, #5514):
//
//	"Flash
//	 When this enchantment enters, it deals 2 damage to any target.
//	 {2}{R}, Sacrifice this enchantment: Scry 2."
//
// Omen of the Sea's frame: flash, an enters trigger, and a sacrifice-
// for-scry-2 activated ability, each Omen paying {2} and one mana of its
// own color. The enters trigger is Ghitu Slinger's 2 damage to any
// target, dealt by the enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e11e2dc8-b28e-415d-9a5d-0eb46da8945c",
		Name:            "Omen of the Forge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Omen of the Forge — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}"), SacrificeThis()),
			Effect:  Do(Scry{N: 2}),
		}},
	})
}
