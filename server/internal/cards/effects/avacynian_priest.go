package effects

// Avacynian Priest — Creature — Human Cleric {1}{W}, 1/2 (Innistrad
// Remastered):
//
//	"{1}, {T}: Tap target non-Human creature."
//
// "Non-Human" is read through the layered view of the target's creature
// types.
func init() {
	Register(Spec{
		OracleID:     "df7e296d-96de-4280-a4d8-49189544b9e3",
		Name:         "Avacynian Priest",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target non-Human creature.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target non-Human creature", Not(HasSubtype("Human"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
