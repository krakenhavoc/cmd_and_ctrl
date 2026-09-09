package effects

// Lotus Petal — Artifact {0} (EDHREC rank 120):
//
//	"{T}, Sacrifice this artifact: Add one mana of any color."
//
// The first CATALOG card to use ManaAbilityCost.Sacrifice, which
// spec.go had reserved for exactly this card since S15 and which
// the engine has honoured since S21 sub-PR 1 (the Treasure token
// exercises the same path). Nothing new is needed to ship it.
//
// The pipe in Produced routes through the same colour-pick prompt
// Birds of Paradise uses, so the mana lands only once the
// controller answers. The autotapper deliberately skips
// sacrifice-cost abilities, so paying a cost never eats the Petal
// on its own.
func init() {
	Register(Spec{
		OracleID: "32e5339e-9e4f-46f8-b305-f9d6d3ba8bb5",
		Name:     "Lotus Petal",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
	})
}
