package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8e720216-a980-4a34-8aef-62ac3ef15a6e",
		Name:         "Leonin Relic-Warder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      effects.Self,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Leonin Relic-Warder — exile target artifact or enchantment?"},
			Targets:        effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Key:            label,
			Effect:         effects.ExileChosenTargetUntilThisLeaves("Leonin Relic-Warder — the exiled card returns when Leonin Relic-Warder leaves the battlefield"),
		}},
	})
}
