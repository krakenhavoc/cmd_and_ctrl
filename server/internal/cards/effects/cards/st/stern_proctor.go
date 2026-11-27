package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stern Proctor — Creature — Human Wizard {U}{U}, 1/2 (Urza's Saga, S82 OLD
// batch 732, #5624):
//
//	"When this creature enters, return target artifact or enchantment to its
//	 owner's hand."
//
// Any artifact or enchantment, anyone's, chosen as the trigger goes on the
// stack and rechecked as it resolves (CR 608.2b); it returns to its owner's
// hand, and a token ceases to exist. With no legal target the trigger is
// removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5120bfdc-7201-420d-85a1-eeb4aba6632b",
		Name:         "Stern Proctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stern Proctor — return target artifact or enchantment to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
		},
	})
}
