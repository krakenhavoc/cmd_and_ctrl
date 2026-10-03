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
// One declared simplification, the paid-cost record's: with strict
// mana off the engine never saw what paid, so no counters (ADR 0068
// §3).
func init() {
	Register(Spec{
		OracleID:     "4fca09ac-8134-43d9-a84b-686db5e2bf69",
		Name:         "Kalain, Reclusive Painter",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't track which mana you spent, so your other creatures get no extra +1/+1 counters."},
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
