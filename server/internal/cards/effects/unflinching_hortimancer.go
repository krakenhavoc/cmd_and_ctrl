package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unflinching Hortimancer — Creature — Human Cleric {1}{W}, 2/1:
//
//	"Ward {1}
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// Ajani's Pridemate's trigger (once per gain event, not per point) under
// a ward {1}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cd0bd74-2463-4e9d-9802-69ac4ef6d543",
		Name:         "Unflinching Hortimancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Unflinching Hortimancer — ward {1}"),
			WheneverYouGainLife("Unflinching Hortimancer — put a +1/+1 counter on it", putCounterOnSelf),
		},
	})
}
