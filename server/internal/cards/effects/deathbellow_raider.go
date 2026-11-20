package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathbellow Raider — Creature — Minotaur Berserker {1}{R}, 2/3 (Theros, #5561):
//
//	"This creature attacks each combat if able.
//	 {2}{B}: Regenerate this creature."
//
// Juggernaut's attack requirement plus the shared regeneration row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b14e7ea-e01c-421d-acd2-0b70436ca6aa",
		Name:         "Deathbellow Raider",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Activated:    []ActivatedAbility{regenerateThisRow("{2}{B}")},
	})
}
