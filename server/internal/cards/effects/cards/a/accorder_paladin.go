package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accorder Paladin — Creature — Human Knight {1}{W}, 3/1 (Mirrodin Besieged,
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
		OracleID:     "6891c8b7-7112-40dc-8b94-b84ad1885e5e",
		Name:         "Accorder Paladin",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BattleCry("Accorder Paladin")},
	})
}
