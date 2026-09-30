package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaea's Protector — Creature — Elemental Warrior, {3}{G}, 4/2:
//
//	"This creature must be blocked if able."
//
// One CR 509.1c requirement on the attacker (#1597): the defending
// player must block it with at least one creature when some legal
// declaration can, and may block it with more. A defender whose
// creatures can't block it owes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d91d5d22-9f94-4a41-9970-0a8ddd2f7711",
		Name:         "Gaea's Protector",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{MustBeBlocked()},
	})
}
