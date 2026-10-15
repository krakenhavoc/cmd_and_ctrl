package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Freestrider Commando — Creature — Centaur Mercenary {2}{G}, 3/3:
//
//	"This creature enters with two +1/+1 counters on it if it wasn't
//	 cast or no mana was spent to cast it.
//	 Plot {3}{G}"
//
// The counters are a CR 614.1c self-replacement that reads the entering
// spell's payment inside the entry window (ADR 0109 §11, #1552,
// game.EntrySpentForEffect): an entry that was not a cast spent nothing,
// and a plotted Commando cast "without paying its mana cost" spent a
// known nothing, so both get the counters. Plot is the ordinary special
// action (CR 702.170).
//
// One declared simplification, the paid-cost record's: when a payment
// is waived the engine sees only the share its pool paid, so a
// Commando cast that way is not known to have spent nothing, and gets
// no counters even when it was cast for free (ADR 0068 §3).
func init() {
	Register(Spec{
		OracleID:       "09117016-5fd1-4590-9867-aad65d3097e3",
		Name:           "Freestrider Commando",
		Completeness:   CompletenessCaveats,
		Caveats:        []string{"When “Charge mana costs” is off, or you choose Cast anyway, the game can't tell a Commando paid outside the mana pool from one cast for no mana, so it doesn't get the two counters — one put onto the battlefield still does."},
		SpecialActions: []game.SpecialAction{Plot("{3}{G}")},
		Replacements: []game.ReplacementEffect{
			SelfEntersWithCountersIfNoManaSpent(game.CounterPlusOne, 2),
		},
	})
}
