package effects

// Sigardian Priest — Creature — Human Cleric {1}{W}, 1/2 (Eldritch Moon,
// #5541):
//
//	"{1}, {T}: Tap target non-Human creature."
//
// Avacynian Priest's ability, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2e22c9d-b0ec-42bb-a520-3c36df546ba0",
		Name:         "Sigardian Priest",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target non-Human creature.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target non-Human creature", Not(HasSubtype("Human"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
