package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Relic-Warder — Creature — Cat Cleric {W}{W}, 2/2:
//
//	"When this creature enters, you may exile target artifact or
//	 enchantment.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// Angel of Sanctions's optional exile-until-leaves (CR 610.3): the exile
// and the return are linked, so the card returns as a new object under its
// owner's control when the Warder leaves, and if the Warder left before the
// trigger resolves nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	const label = "Leonin Relic-Warder — you may exile target artifact or enchantment"
	Register(Spec{
		OracleID:     "8e720216-a980-4a34-8aef-62ac3ef15a6e",
		Name:         "Leonin Relic-Warder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      Self,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Leonin Relic-Warder — exile target artifact or enchantment?"},
			Targets:        TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Key:            label,
			Effect:         exileChosenTargetUntilThisLeaves("Leonin Relic-Warder — the exiled card returns when Leonin Relic-Warder leaves the battlefield"),
		}},
	})
}
