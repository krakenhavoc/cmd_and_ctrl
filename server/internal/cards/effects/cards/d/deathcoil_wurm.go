package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathcoil Wurm — Creature — Wurm {6}{G}{G}, 7/6 (Portal Second Age, S82
// older sets batch 733, #5625):
//
//	"You may have this creature assign its combat damage as though it
//	 weren't blocked."
//
// Lone Wolf's layer 6 flag (CR 510.1c); the "may" is the engine's
// per-creature choice as the blocked creature assigns damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48408f60-56bc-402f-ba19-c9dc53836765",
		Name:         "Deathcoil Wurm",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB614AssignAsThoughUnblocked()},
	})
}
