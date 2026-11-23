package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f21c3d22-6149-475f-ae6e-37a25ef020e5",
		Name:         "Gaea's Herald",
		Completeness: effects.CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.AnyPlayersSpellsCantBeCountered("Creature spells can't be countered.", effects.Creature()),
		},
	})
}
