package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Partisan — Creature — Elephant Soldier {4}{W}, 3/4 (Mirrodin Besieged,
// #5576):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)"
//
// Battle cry is CR 702.91a's trigger row (BattleCry, battle_cry.go): each
// other creature attacking as it resolves gets +1/+0 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d2bfbb2-84df-4d7c-9989-7f5577f90d97",
		Name:         "Loxodon Partisan",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BattleCry("Loxodon Partisan")},
	})
}
