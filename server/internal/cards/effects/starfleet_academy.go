package effects

// Starfleet Academy — Legendary Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 Federation — {5}, {T}: Choose a color. Add an amount of mana of that
//	 color equal to the number of creature types among non-Borg
//	 creatures you control."
//
// Federation is printed as an ability word (CR 207.2c): it has no rules
// meaning. The second
// ability is Nykthos's shape — a {5}-and-tap mana ability whose one colour
// pick adds N of that colour (ProducedOneColor), N counted as it is
// activated: each creature type among the controller's non-Borg
// creatures once (CR 205.3m), through the layers, a changeling counting
// as a Borg (CR 702.73a). With no such creature it adds nothing and the
// land still taps, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de8ebef9-fa0e-42df-a1bc-5c4b86d1bdd1",
		Name:         "Starfleet Academy",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:           ManaAbilityCost{Tap: true, Mana: "{5}"},
				ProducedAmount: ProducedOneColor(trekB29StarfleetAcademyAmount),
				Label:          "Federation — Choose a color: add that much mana for each creature type among non-Borg creatures you control",
			},
		},
	})
}
