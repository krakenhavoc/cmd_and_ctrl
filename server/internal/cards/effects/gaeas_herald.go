package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaea's Herald — Creature — Elf {1}{G}, 1/1:
//
//	"Creature spells can't be countered."
//
// ADR 0106 §4's battlefield static (#1806), in its "any player" form:
// there is no "you", so every player's creature spells are covered,
// opponents' included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f21c3d22-6149-475f-ae6e-37a25ef020e5",
		Name:         "Gaea's Herald",
		Completeness: CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			AnyPlayersSpellsCantBeCountered("Creature spells can't be countered.", Creature()),
		},
	})
}
