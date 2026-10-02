package effects

// Hurr Jackal — Creature — Jackal {R}, 1/1:
//
//	"{T}: Target creature can't be regenerated this turn."
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d17f5afa-a884-4b99-aa9e-89ddb3d43b22",
		Name:         "Hurr Jackal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature can't be regenerated this turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeRegeneratedThisTurn,
		}},
	})
}
