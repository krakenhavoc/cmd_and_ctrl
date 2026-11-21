package effects

// Icy Manipulator — Artifact {4} (Dominaria Remastered):
//
//	"{1}, {T}: Tap target artifact, creature, or land."
//
// Avacynian Priest's tap ability over a wider target clause: any
// artifact, creature or land, whoever controls it. A permanent that is
// already tapped is a legal target and stays tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3608f1f7-8dc5-4dd1-ae91-c830e1de9529",
		Name:         "Icy Manipulator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target artifact, creature, or land.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
			Effect:  tapChosenPermanent,
		}},
	})
}
