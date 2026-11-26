package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suspension Field — Enchantment {1}{W} (Khans of Tarkir):
//
//	"When this enchantment enters, you may exile target creature with toughness
//	 3 or greater until this enchantment leaves the battlefield. (That creature
//	 returns under its owner's control.)"
//
// Mystical Tether's exile, with the target chosen as the trigger goes on the
// stack (CR 603.3d) and the "may" answered as it resolves; the card returns
// when the enchantment leaves (CR 610.3), and is not exiled at all if the
// enchantment has already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "baf8310f-5baf-459e-95ae-6b43a1d47e24",
		Name:         "Suspension Field",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Suspension Field — you may exile target creature with toughness 3 or greater until this enchantment leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Suspension Field — the exiled card returns when Suspension Field leaves the battlefield")),
				effects.TargetCreature("target creature with toughness 3 or greater", effects.ToughnessGE(3))),
				"Suspension Field — exile the target creature until this enchantment leaves the battlefield?"),
		},
	})
}
