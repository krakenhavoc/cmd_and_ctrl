package effects

// Atzocan Seer — Creature — Human Druid {1}{G}{W}, 2/3 (Lost Caverns of
// Ixalan Commander):
//
//	"{T}: Add one mana of any color.
//	 Sacrifice this creature: Return target Dinosaur card from your
//	 graveyard to your hand."
//
// The sacrifice is a cost (CR 118.3) paid before anything resolves, and
// the target is a Dinosaur card in the graveyard as the ability goes on the stack.
func init() {
	Register(Spec{
		OracleID:     "affced7f-797d-4e50-8a24-2da14de44e43",
		Name:         "Atzocan Seer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Return target Dinosaur card from your graveyard to your hand.",
			Cost:    SacrificeThis(),
			Targets: TargetCardInGraveyard("target Dinosaur card from your graveyard", YouOwn(), Subtype("Dinosaur")),
			Effect:  returnTargetedCardToHand,
		}},
	})
}
