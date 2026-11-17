package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1612a08f-fc1f-45cf-ae03-5f021a46869d",
		Name:         "In the Trenches",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16Anthem(b16CreaturesYouControl, 1, 1)},
		Activated: []ActivatedAbility{{
			Label: "{5}{W}: Exile target nonland permanent you don't control until this enchantment leaves the battlefield. " +
				"Activate only as a sorcery and only once.",
			Cost:         ManaCost("{5}{W}"),
			SorcerySpeed: true,
			OnlyOnce:     true,
			Targets:      TargetPermanent("target nonland permanent you don't control", And(Nonland(), Not(YouControl()))),
			Effect:       exileChosenTargetUntilThisLeaves("In the Trenches — returns when In the Trenches leaves the battlefield"),
		}},
	})
}
