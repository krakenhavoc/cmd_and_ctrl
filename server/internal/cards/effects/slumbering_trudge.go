package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slumbering Trudge — Creature — Plant Beast {X}{G}, 6/6 (Secrets of
// Strixhaven, #4456):
//
//	"This creature enters with a number of stun counters on it equal to three
//	 minus X. If X is 2 or less, it enters tapped. (If a permanent with a stun
//	 counter would become untapped, remove one from it instead.)"
//
// Both sentences read the X the spell was cast with, as the permanent enters
// (CR 614.1c, CR 107.3m): the counters are an EntersWithCountersFromCast
// clause (three minus X, never below none) and the tapped entry is a
// self-replacement over the same X (socB34EntersTappedIfXAtMost). They are
// separate hooks, so no CR 616 ordering prompt opens. The stun counter's
// untap-replacement is the engine's (CR 122.1g). A Trudge put onto the
// battlefield without being cast has X = 0 (CR 107.3m): three stun counters,
// tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "d04bce9a-822b-4a52-8826-b503ef5f5b2e",
		Name:                       "Slumbering Trudge",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{socB34StunCountersThreeMinusX()},
		Replacements:               []game.ReplacementEffect{socB34EntersTappedIfXAtMost(2)},
	})
}
