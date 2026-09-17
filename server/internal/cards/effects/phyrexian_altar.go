package effects

// Phyrexian Altar — Artifact {3}:
//
//	"Sacrifice a creature: Add one mana of any color."
//
// Ashnod's Altar's coloured cousin: one mana instead of two, but any
// colour, which is what makes it the combo piece of choice when the
// loop needs a specific colour rather than raw quantity.
//
// The five-colour pipe offers all five colours, the controller's
// commander identity listed first, as Treasure and Birds of Paradise
// do.
func init() {
	Register(Spec{
		OracleID: "8d02b297-97c4-4379-9862-0a462400f66f",
		Name:     "Phyrexian Altar",
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				SacrificeOther: SacrificeACreature().SacrificeOther,
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Sacrifice a creature: Add one mana of any color",
		}},
	})
}
