package effects

// Mockingbird — Creature — Bird Bard {X}{U}, 1/1:
//
//	"Flying
//	 You may have this creature enter as a copy of any creature on the
//	 battlefield with mana value less than or equal to the amount of
//	 mana spent to cast this creature, except it's a Bird in addition
//	 to its other types and it has flying."
//
// Flying is its own printed line — Mockingbird's baseline body when
// it doesn't copy anything, or can't.
//
// Declared simplification: the copy-choice half is dropped entirely.
// "The amount of mana spent to cast this creature" needs a read of
// the resolving spell's PAID COST from inside the CR 614 entry
// window, and the engine deliberately does not expose one there.
// `entry_counters.go`'s own doc says so: `CastCounts` is the one,
// narrow, documented channel a catalog entry point may read at that
// moment, and it carries only X, kicker count and colours of mana
// spent (CR 614.1c's needs) — nothing for "how much mana, total".
// `CopySelector.Candidates`, the hook this card would need, isn't
// even handed a `CastCounts` — it gets only the game, the controller
// and the entering card's own ID. Reaching past that into
// `StackItemPaidForEffect` doesn't help either: the stack item behind
// an entering permanent is passed down the resolution call chain by
// value at that point (see `applyCastEntryCountersLocked`'s `item`
// parameter) rather than being re-derivable from the public accessor,
// so a card-side lookup by ID there is not guaranteed to find
// anything and would silently answer "zero mana spent" — a ceiling of
// mana value 0, which is not an honest approximation of the printed
// card, just a broken one dressed as a number.
//
// Shipping the whole clause out leaves a plain {X}{U} 1/1 flier,
// which is unambiguously weaker than printed and never stronger — the
// #259 direction. It becomes writable once the entry window carries a
// "total mana spent" figure the way it already carries X and
// ColorsSpent.
func init() {
	Register(Spec{
		OracleID:        "b9df2cdf-397c-458d-89cf-911568737ffa",
		Name:            "Mockingbird",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"flying"},
		Caveats: []string{
			"It can't enter as a copy of another creature — the entry window it would need to read \"the amount of mana spent to cast it\" from doesn't expose that figure. It's just an {X}{U} 1/1 flier.",
		},
	})
}
