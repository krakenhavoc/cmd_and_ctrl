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
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) no longer leaves this card guessing: it declares ReadsManaSpent,
// so the caster is asked what they paid outside the pool, or whether they
// cast it without paying its mana cost, and the record it reads is exact
// (#3855, ADR 0068 amendment 2026-10-15).
// A Commando declared cast without paying its mana cost, with nothing
// from the pool, gets its counters as printed.
func init() {
	Register(Spec{
		OracleID:       "09117016-5fd1-4590-9867-aad65d3097e3",
		Name:           "Freestrider Commando",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{3}{G}")},
		Replacements: []game.ReplacementEffect{
			SelfEntersWithCountersIfNoManaSpent(game.CounterPlusOne, 2),
		},
	})
}
