package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thelon's Chant — Enchantment {1}{G}{G} (Fallen Empires, #5746):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless
//	 you pay {G}.
//	 Whenever a player puts a Swamp onto the battlefield, this enchantment
//	 deals 3 damage to that player unless the player puts a -1/-1 counter
//	 on a creature they control."
//
// The upkeep is UpkeepSacrificeUnlessYouPay, the shared pay-or-sacrifice
// shape. The Swamp trigger is the Chants' punisher (chantLandPunisher,
// minus_counter_payment.go): the player who PUT the Swamp there (the
// event's PutBy, even if it entered under someone else, per the
// 2004-10-04 ruling) is asked to put a -1/-1 counter on a creature they
// control as a pay-unless payment (game.PayActionMinusOneCounter), or
// take 3 damage. With no creature that could take the counter they are
// not asked and take the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff4ccf8d-7d3d-4a44-a9ab-6bc88e90d6a7",
		Name:         "Thelon's Chant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Thelon's Chant", "{G}"),
			chantLandPunisher("Thelon's Chant", "Swamp"),
		},
	})
}
