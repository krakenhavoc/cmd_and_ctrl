package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathbellow Raider — Creature — Minotaur Berserker {1}{R}, 2/3 (Theros, #5561):
//
//	"This creature attacks each combat if able.
//	 {2}{B}: Regenerate this creature."
//
// Juggernaut's attack requirement plus the shared regeneration row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b14e7ea-e01c-421d-acd2-0b70436ca6aa",
		Name:         "Deathbellow Raider",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{2}{B}")},
	})
}
