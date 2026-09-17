package effects

// Cultivator's Caravan — Artifact — Vehicle, 5/5, for {3}:
//
//	"{T}: Add one mana of any color.
//	 Crew 3"
//
// The mana ability and the crew ability are both on the card at
// once, and they compete: tapping for mana leaves the Caravan tapped,
// and a crewed Caravan that has already tapped for mana cannot
// attack. That is the printed card, and it falls out for free —
// nothing here has to encode the tension.
//
// The five-colour pipe offers all five colours like every other "any
// color" source, with the commander's identity listed first, so a
// mono-green deck sees {G} as the first button of five.
func init() {
	Register(Spec{
		OracleID:     "c1eb530c-dd36-40ae-8617-6bb6969565e1",
		Name:         "Cultivator's Caravan",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("Cultivator's Caravan"),
		}},
	})
}
