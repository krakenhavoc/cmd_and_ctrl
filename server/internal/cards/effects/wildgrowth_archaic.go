package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildgrowth Archaic — Creature — Avatar {2/G}{2/G}, 0/0 (Secrets of
// Strixhaven, #4456):
//
//	"Reach, trample
//	 Converge — This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.
//	 Whenever you cast a creature spell, that creature enters with X
//	 additional +1/+1 counters on it, where X is the number of colors of
//	 mana spent to cast it."
//
// Converge is an "enters with" clause read off its own payment
// (CountersPerColorSpent): two green paid for both hybrid symbols is one
// colour and one counter; generic paid for them is none, and the 0/0 dies.
//
// The cast trigger reads X off the creature spell's payment as it is
// cast and, on resolution, marks that spell on the stack (#4686): the
// counters are part of its entry, so Hardened Scales and Doubling Season
// see them. A countered trigger, or a creature spell that has left the
// stack, gives nothing; a copy of the spell was not cast and gets none.
// Both abilities read mana spent, so a waived payment asks what was paid
// outside the pool, and the auto-tapper prefers distinct colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "4469ff35-54ec-4ff5-bc19-3808ae0f711b",
		Name:                "Wildgrowth Archaic",
		ReadsManaSpent:      game.ReadsOwnPayment | game.ReadsYourCasts,
		Completeness:        CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{"reach", "trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerColorSpent(game.CounterPlusOne),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCastThatEntersWithCounters("Wildgrowth Archaic", Creature(), game.CounterPlusOne, colorsOfManaSpent,
				"Wildgrowth Archaic — that creature enters with an additional +1/+1 counter for each color of mana spent to cast it"),
		},
	})
}
