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
// The five-colour pipe goes through the commander-identity filter
// like every other "any color" source, so a mono-green deck is
// offered {G} rather than a five-way prompt. That narrowing is the
// engine's default and is right here: the printed text says "any
// color", but no mana outside the commander's identity can legally
// be spent in this format anyway.
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
