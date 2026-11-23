package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e11e2dc8-b28e-415d-9a5d-0eb46da8945c",
		Name:            "Omen of the Forge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Omen of the Forge — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
