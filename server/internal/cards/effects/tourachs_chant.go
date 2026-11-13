package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tourach's Chant — Enchantment {1}{B}{B} (Fallen Empires, #5746):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless
//	 you pay {B}.
//	 Whenever a player puts a Forest onto the battlefield, this
//	 enchantment deals 3 damage to that player unless they put a -1/-1
//	 counter on a creature they control."
//
// Thelon's Chant with the colours swapped: the upkeep is
// UpkeepSacrificeUnlessYouPay, and the Forest trigger the Chants' punisher
// (chantLandPunisher, minus_counter_payment.go). The player who put the
// Forest there chooses between a -1/-1 counter on a creature they control
// (game.PayActionMinusOneCounter) and 3 damage; with no creature that
// could take the counter they take the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdf4e225-23b5-47d9-9936-ddcec8aa1ca4",
		Name:         "Tourach's Chant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Tourach's Chant", "{B}"),
			chantLandPunisher("Tourach's Chant", "Forest"),
		},
	})
}
