package effects

// Mudflat Village — Land (Bloomburrow, #4467):
//
//	"{T}: Add {C}.
//	 {T}: Add {B}. Spend this mana only to cast a creature spell.
//	 {1}{B}, {T}, Sacrifice this land: Return target Bat, Lizard, Rat,
//	 or Squirrel card from your graveyard to your hand."
//
// Bloomburrow's black creature-type land, built like Lilypad Village:
// the {B} carries the cast and creature restriction tags, and is a
// deliberate click since the auto-tapper leaves restricted abilities
// alone. The sacrifice ability is Buried Ruin's, with the target
// narrowed to a card with one of the four subtypes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aeeab1df-0b8b-4bc4-a5f9-aac413449bec",
		Name:         "Mudflat Village",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B} (creature spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, {T}, Sacrifice this land: Return target Bat, Lizard, Rat, or Squirrel card from your graveyard to your hand.",
			Cost:  Plus(ManaCost("{1}{B}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target Bat, Lizard, Rat, or Squirrel card in your graveyard",
				AnySubtype("Bat", "Lizard", "Rat", "Squirrel"), YouOwn()),
			Effect: returnTargetedCardToHand,
		}},
	})
}
