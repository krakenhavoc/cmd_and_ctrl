package effectstest

// sneak_test.go — #4586: sneak (CR 702.190) and its first card, Elektra,
// Daughter of the Hand. The engine half is game/sneak.go; the cost is
// ninjutsu's unblocked attacker (ninjutsu_test.go) paid as an
// alternative cost (Daze's ReturnToHand).

const (
	ElektraDaughterOracle = "b46647a1-0c91-451b-8c0b-be8fe02faee2"
	ElektraDaughterName   = "Elektra, Daughter of the Hand"
	ElektraDaughterType   = "Legendary Creature — Human Ninja Villain"
	// A test-only instant with sneak: no printed instant with the keyword
	// is in the catalog yet (TMNT, #4457, has them).
	SneakInstantOracle = "test-sneak-instant-4586"
)
