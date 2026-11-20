package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorch Rider — Creature — Human Warrior {3}{R}, 4/3 (Zendikar Rising,
// #5508):
//
//	"Kicker {1}{R} (You may pay an additional {1}{R} as you cast this
//	 spell.)
//	 When this creature enters, if it was kicked, it gains haste until
//	 end of turn."
//
// Goblin Bushwhacker's kicked enters trigger, an intervening if (CR
// 603.4) read off the record the cast carried onto the permanent, so a
// Rider put onto the battlefield without being cast was not kicked. The
// grant is thisCreatureUntilEOT, which does nothing if the Rider left
// in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8165acd4-ac62-47f1-97dd-01860aee29ee",
		Name:          "Scorch Rider",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Scorch Rider — kicked, it gains haste until end of turn",
				thisCreatureUntilEOT("Scorch Rider — haste", 0, 0, "haste")),
		},
	})
}
