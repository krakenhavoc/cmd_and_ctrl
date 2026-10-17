package effects

// Master's Rebuke — Instant {1}{G}:
//
//	"Target creature you control deals damage equal to its power to
//	 target creature or planeswalker you don't control."
//
// Bite Down's text word for word: two target clauses, slot 0 the source
// (its power read as the spell resolves, deathtouch and lifelink carried)
// and slot 1 the recipient. bitingSpell is the shared spec and
// oneSidedBite the shared resolution.
//
// No simplification.
func init() {
	Register(bitingSpell("e68ff836-6e51-4c3a-91e5-f0d649297245", "Master's Rebuke"))
}
