package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightwielder Paladin — Creature — Human Knight {3}{W}{W}, 4/4 (Magic 2010, #5585):
//
//	"First strike
//	 Whenever this creature deals combat damage to a player, you may exile
//	 target black or red permanent that player controls."
//
// The target is chosen as the trigger is put on the stack, among the
// permanents of the player that was dealt damage (Sunder Shaman's shape); the
// "may" is asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad62e9d6-61bc-485f-a9b8-2ef9ba98a4b0",
		Name:            "Lightwielder Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.WheneverThisDealsCombatDamageToAPlayer("Lightwielder Paladin — exile target black or red permanent that player controls",
					effects.ExileFirstTarget), "Exile target black or red permanent that player controls?"),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target black or red permanent that player controls", effects.Or(effects.OfColor("B"), effects.OfColor("R"))))),
		},
	})
}
