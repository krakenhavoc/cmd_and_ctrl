package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kumano's Pupils — Creature — Human Shaman {4}{R}, 3/3:
//
//	"If a creature dealt damage by this creature this turn would die,
//	 exile it instead."
//
// Kumano, Master Yamabushi's replacement (#1946,
// ExileIfDealtDamageByThisWouldDie), reached here through combat damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7f8a186-313e-4a29-bf4e-b6a200dfd1ed",
		Name:         "Kumano's Pupils",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			ExileIfDealtDamageByThisWouldDie("Kumano's Pupils: exile it instead of dying"),
		},
	})
}
