package effects

// Excruciator — Creature — Avatar {6}{R}{R}, 7/7:
//
//	"Damage that would be dealt by this creature can't be prevented."
//
// A static about its own damage (UnpreventableByThis, ADR 0107 §5),
// read as each damage event opens — combat damage, a fight — and from
// last-known information once it has left the battlefield (CR 608.2h).
// Protection doesn't stop it either: protection's damage clause is a
// prevention effect (CR 702.16e).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:              "9911c2f8-fafb-4979-898e-e86b726e3538",
		Name:                  "Excruciator",
		Completeness:          CompletenessFull,
		DamageCantBePrevented: DamageByThisCantBePrevented(),
	})
}
