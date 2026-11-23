package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kuldotha Ringleader — Creature — Giant Berserker {4}{R}, 4/4 (Mirrodin Besieged,
// #5576):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 This creature attacks each combat if able."
//
// Battle cry (BattleCry, CR 702.91a) and Juggernaut's attack requirement
// (AttacksEachCombat, CR 508.1d): the declaration is refused while it could
// attack and doesn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dff72513-b67a-4732-a8fc-269cbf21ae99",
		Name:         "Kuldotha Ringleader",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered:    []game.TriggeredAbility{effects.BattleCry("Kuldotha Ringleader")},
	})
}
