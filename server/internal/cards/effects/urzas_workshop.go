package effects

// Urza's Workshop — Land — Urza's (The Brothers' War):
//
//	"{T}: Add {C}.
//	 Metalcraft — {T}: Add {C} for each Urza's land you control. Activate
//	 only if you control three or more artifacts."
//
// Two mana abilities (CR 605.1a), both tapping the land. The metalcraft
// one is Cabal Coffers' scaled shape (ProducedPerPermanent over the
// effective land subtype "Urza's", so Workshop counts itself and every
// Tron land or Urza's Factory) with Mox Opal's three-artifact gate
// (ControlsAtLeast(3, MatchArtifact)). "Metalcraft —" is an ability word
// (CR 207.2c): the activation condition is the whole of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71099427-e110-488f-ab29-7867241fc7f0",
		Name:         "Urza's Workshop",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:           ManaAbilityCost{Tap: true},
				ProducedAmount: ProducedPerPermanent("C", MatchLandSubtype("Urza's")),
				Label:          "Metalcraft — add {C} for each Urza's land you control",
				Condition:      ControlsAtLeast(3, MatchArtifact),
			},
		},
	})
}
