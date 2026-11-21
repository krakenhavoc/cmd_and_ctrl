package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Wardriver — Creature — Goblin Warrior {R}{R}, 2/2 (Mirrodin Besieged,
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
		OracleID:     "10849bbb-50ca-4194-a7cb-fa872272b29b",
		Name:         "Goblin Wardriver",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BattleCry("Goblin Wardriver")},
	})
}
