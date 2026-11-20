package effects

// Ninja of the New Moon — Creature — Spirit Ninja {3}{B}{B}, 6/3 (Modern Horizons, #5519):
//
//	"Ninjutsu {3}{B} ({3}{B}, Return an unblocked attacker you control to
//	 hand: Put this card onto the battlefield from your hand tapped and
//	 attacking.)"
//
// Ninjutsu is the shared hand ability (CR 702.49), Dokuchi Shadow-Walker's
// cost on a 6/3 with nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf44b9b7-180c-4d8c-a410-9561e4c8fb9b",
		Name:         "Ninja of the New Moon",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{3}{B}")},
	})
}
