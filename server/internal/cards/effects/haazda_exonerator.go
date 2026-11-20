package effects

// Haazda Exonerator — Creature — Human Cleric {W}, 1/1 (Dissension, #5598):
//
//	"{T}, Sacrifice this creature: Destroy target Aura."
//
// Goblin Digging Team's row with an Aura for the Wall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6df2e8c9-d87e-4f0d-a3d3-b8eee5cdafd0",
		Name:         "Haazda Exonerator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target Aura.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target Aura", HasSubtype("Aura")),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
