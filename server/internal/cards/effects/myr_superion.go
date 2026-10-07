package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myr Superion — Artifact Creature — Myr {2}, 5/6:
//
//	"Spend only mana produced by creatures to cast this spell."
//
// #2556: the restriction is on the mana's SOURCE (SpendOnlySources, ADR
// 0040's 2026-10-07 amendment). A creature's mana is a Llanowar Elves'
// or a Myr token's; a land's is not, and the auto-tapper plans from
// creatures alone. A land that is also a creature (Dryad Arbor, an
// animated Mutavault) counts while it is one, because the kinds are
// read as the mana is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "43dac6fa-4cdf-49ec-9e11-2e5e8f9f928b",
		Name:             "Myr Superion",
		Completeness:     CompletenessFull,
		SpendOnlySources: game.ManaSourceCreature,
	})
}
