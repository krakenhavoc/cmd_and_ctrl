package effects

// Magus of the Arena — Creature — Human Wizard {4}{R}{R}, 5/5:
//
//	"{3}, {T}: Tap target creature you control and target creature of
//	 an opponent's choice they control. Those creatures fight each
//	 other."
//
// Arena's ability on a creature (arena.go has the rulings): the second
// target is an opponent's choice (#2055), the Magus's controller always
// chooses their creature first, and a retarget keeps the chosen
// opponent. The Magus may be its own first target. {T} is a creature's
// tap, so summoning sickness applies (CR 302.6).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "44865261-16f8-42d2-a388-a57173142eb0",
		Name:         "Magus of the Arena",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{arenaRow()},
	})
}
