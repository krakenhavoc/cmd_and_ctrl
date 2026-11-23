package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// In the Trenches — Enchantment {1}{W}{W} (The Brothers' War):
//
//	"Creatures you control get +1/+1.
//	 {5}{W}: Exile target nonland permanent you don't control until this
//	 enchantment leaves the battlefield. Activate only as a sorcery and only
//	 once."
//
// The exile is the CR 610.3 "until" shape (Angel of Condemnation's), so the
// card returns the moment this leaves, and nothing is exiled at all if this
// had already left when the ability resolves. "Only once" is the shared
// OnlyOnce gate, counted per permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1612a08f-fc1f-45cf-ae03-5f021a46869d",
		Name:         "In the Trenches",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B16CreaturesYouControl, 1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label: "{5}{W}: Exile target nonland permanent you don't control until this enchantment leaves the battlefield. " +
				"Activate only as a sorcery and only once.",
			Cost:         effects.ManaCost("{5}{W}"),
			SorcerySpeed: true,
			OnlyOnce:     true,
			Targets:      effects.TargetPermanent("target nonland permanent you don't control", effects.And(effects.Nonland(), effects.Not(effects.YouControl()))),
			Effect:       effects.ExileChosenTargetUntilThisLeaves("In the Trenches — returns when In the Trenches leaves the battlefield"),
		}},
	})
}
