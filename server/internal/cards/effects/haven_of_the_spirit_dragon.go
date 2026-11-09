package effects

// Haven of the Spirit Dragon — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Dragon
//	 creature spell.
//	 {2}, {T}, Sacrifice this land: Return target Dragon creature card or
//	 Ugin planeswalker card from your graveyard to your hand."
//
// Beastcaller Savant's restricted any-colour mana, restricted to Dragon
// creature spells (Carnelian Orb's restriction triple); the mana pays any
// part of such a spell's cost, additional or alternative included. The
// return is Memorial to Folly's, aimed at either kind of card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acc9c16a-5e72-43bd-87e1-56a16aa892f5",
		Name:         "Haven of the Spirit Dragon",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype("Dragon")},
				Label:        "Add one mana of any color. Spend this mana only to cast a Dragon creature spell",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this land: Return target Dragon creature card or Ugin planeswalker card from your graveyard to your hand.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target Dragon creature card or Ugin planeswalker card in your graveyard",
				YouOwn(), Or(And(Creature(), Subtype("Dragon")), And(Planeswalker(), Subtype("Ugin")))),
			Effect: b27ReturnChosenGraveyardCardToHand,
		}},
	})
}
