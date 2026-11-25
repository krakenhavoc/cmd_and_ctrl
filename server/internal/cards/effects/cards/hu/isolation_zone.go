package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isolation Zone — Enchantment {2}{W}{W} (Oath of the Gatewatch, #5545):
//
//	"When this enchantment enters, exile target creature or enchantment an
//	 opponent controls until this enchantment leaves the battlefield. (That
//	 permanent returns under its owner's control.)"
//
// Mystical Tether's exile (CR 610.3): the target is chosen as the trigger goes
// on the stack and returns when Isolation Zone leaves the battlefield; if it
// has already left by the time the trigger resolves, the permanent is not
// exiled at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c1890cf-36db-4720-a472-9158bdc0d8aa",
		Name:         "Isolation Zone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Isolation Zone — exile target creature or enchantment an opponent controls until this enchantment leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Isolation Zone — the exiled permanent returns when Isolation Zone leaves the battlefield")),
				effects.TargetPermanent("target creature or enchantment an opponent controls", effects.Or(effects.Creature(), effects.Enchantment()), effects.OpponentControls())),
		},
	})
}
