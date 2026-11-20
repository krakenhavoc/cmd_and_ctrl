package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6891c8b7-7112-40dc-8b94-b84ad1885e5e",
		Name:         "Accorder Paladin",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{BattleCry("Accorder Paladin")},
	})
}
