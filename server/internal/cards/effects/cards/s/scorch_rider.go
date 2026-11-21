package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "8165acd4-ac62-47f1-97dd-01860aee29ee",
		Name:          "Scorch Rider",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Scorch Rider — kicked, it gains haste until end of turn",
				effects.ThisCreatureUntilEOT("Scorch Rider — haste", 0, 0, "haste")),
		},
	})
}
