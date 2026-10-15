package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kalain, Reclusive Painter — Legendary Creature — Human Elf Bard
// {B}{R}, 1/2:
//
//	"When Kalain enters, create a Treasure token.
//	 Other creatures you control enter with an additional +1/+1
//	 counter on them for each mana from a Treasure spent to cast them."
//
// The second ability is a CR 614.1c replacement on Kalain that applies
// to another creature's entry (ADR 0109 §11, #1552): it reads the
// entering spell's payment inside the entry window
// (game.EntrySpentForEffect), and each mana remembers that its source
// was a Treasure from the moment it was made (#1212), so the Treasure
// sacrificed to pay still counts. Kalain itself is not "other", and a
// creature that was not cast spent nothing and gets nothing.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, and
// Treasure mana exists only as mana in the pool, so the record says
// exactly whether (and how much) Treasure mana paid.
func init() {
	Register(Spec{
		OracleID:     "4fca09ac-8134-43d9-a84b-686db5e2bf69",
		Name:         "Kalain, Reclusive Painter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kalain, Reclusive Painter — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Replacements: []game.ReplacementEffect{
			CreaturesYouControlEnterWithCountersPerManaFrom(
				"Kalain, Reclusive Painter: a +1/+1 counter for each mana from a Treasure", game.ManaSourceTreasure, true),
		},
	})
}
