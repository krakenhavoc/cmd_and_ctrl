package effects

// Animal Sanctuary — Land (EDHREC rank 3770):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Put a +1/+1 counter on target Bird, Cat, Dog, Goat, Ox,
//	 or Snake."
//
// The pet-typal utility land. The colourless mana is an ordinary tap
// ability; the counter is a real activated ability with a target
// clause over permanents carrying any of the six creature types —
// read post-layer, so a changeling qualifies — and the counter goes
// through the CR 614 placement pipeline, so a Hardened Scales sees
// it. Anyone's animal is a legal target, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3c40943-1d7c-4ea2-b34f-8df8b6775701",
		Name:         "Animal Sanctuary",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on target Bird, Cat, Dog, Goat, Ox, or Snake.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPermanent("target Bird, Cat, Dog, Goat, Ox, or Snake", AnySubtype("Bird", "Cat", "Dog", "Goat", "Ox", "Snake")),
			Effect:  b36CounterOnChosenAnimal,
		}},
	})
}
