package effects

// Furnace Brood — Creature — Elemental {3}{R}, 3/3:
//
//	"{R}: Target creature can't be regenerated this turn."
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cc87a268-253b-4377-bdc5-474940de8878",
		Name:         "Furnace Brood",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: Target creature can't be regenerated this turn.",
			Cost:    ManaCost("{R}"),
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeRegeneratedThisTurn,
		}},
	})
}
