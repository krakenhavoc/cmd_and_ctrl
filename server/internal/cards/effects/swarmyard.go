package effects

// Swarmyard — Land:
//
//	"{T}: Add {C}.
//	 {T}: Regenerate target Insect, Rat, Spider, or Squirrel."
//
// Asceticism's regeneration shield with a creature-type clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4b508087-99da-4eb1-8b12-29162f2ec85d",
		Name:          "Swarmyard",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{T}: Regenerate target Insect, Rat, Spider, or Squirrel.",
			Cost:  TapCost(),
			Targets: TargetCreature("target Insect, Rat, Spider, or Squirrel",
				Or(HasSubtype("Insect"), HasSubtype("Rat"), HasSubtype("Spider"), HasSubtype("Squirrel"))),
			Effect: regenerateTheTargetPermanent,
		}},
	})
}
