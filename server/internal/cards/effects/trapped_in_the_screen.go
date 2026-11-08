package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trapped in the Screen — Enchantment {2}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"Ward {2} (Whenever this enchantment becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {2}.)
//	 When this enchantment enters, exile target artifact, creature, or
//	 enchantment an opponent controls until this enchantment leaves the
//	 battlefield."
//
// Ward {2} is the shared Ward trigger (CR 702.21a). The exile is Glass
// Casket's "until this leaves the battlefield" (CR 610.3) with Banishing
// Light's opponent clause: the card returns under its owner's control the
// moment the enchantment leaves, and an enchantment that leaves before
// its trigger resolves exiles nothing (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20f2fc0e-b1d4-4f80-9c7a-a9fb25fef7b8",
		Name:         "Trapped in the Screen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Trapped in the Screen — ward {2}"),
			Targeting(WhenThisEnters("Trapped in the Screen — exile target artifact, creature, or enchantment an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Trapped in the Screen — the exiled card returns when Trapped in the Screen leaves the battlefield")),
				TargetPermanent("target artifact, creature, or enchantment an opponent controls",
					Or(Artifact(), Creature(), Enchantment()), OpponentControls())),
		},
	})
}
