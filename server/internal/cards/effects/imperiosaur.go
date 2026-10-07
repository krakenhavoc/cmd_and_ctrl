package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperiosaur — Creature — Dinosaur {2}{G}{G}, 5/5:
//
//	"Spend only mana produced by basic lands to cast this spell."
//
// #2556: the restriction is on the mana's SOURCE, so it rides the cast's
// spend context (SpendOnlySources, ADR 0040's 2026-10-07 amendment): the
// pool refuses mana whose recorded source is not a basic land, and the
// auto-tapper plans from basic lands alone, so a Command Tower, a Sol
// Ring or a Treasure never pays for it. Mana that already floats from a
// Sol Ring does not count, nor does mana from a spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e9ced5d8-8337-403f-86a3-bddb9c77d658",
		Name:             "Imperiosaur",
		Completeness:     CompletenessFull,
		SpendOnlySources: game.ManaSourceBasicLand,
	})
}
