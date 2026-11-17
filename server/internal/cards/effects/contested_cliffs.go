package effects

// Contested Cliffs — Land (Onslaught):
//
//	"{T}: Add {C}.
//	 {R}{G}, {T}: Target Beast creature you control fights target
//	 creature an opponent controls."
//
// Ulvenwald Tracker's two-clause fight (slotZeroFightsSlotOne) on a
// land: slot 0 is a Beast you control, slot 1 a creature an opponent
// controls. If either target is illegal as the ability resolves,
// neither fights (CR 701.14b). A land has no summoning sickness, so the
// {T} may be paid the turn it is played.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b891a683-2ebc-4e9c-b402-5dd9c1b42b69",
		Name:          "Contested Cliffs",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{R}{G}, {T}: Target Beast creature you control fights target creature an opponent controls.",
			Cost:  Plus(ManaCost("{R}{G}"), TapCost()),
			Targets: Clauses(
				TargetCreature("target Beast creature you control", YouControl(), OfCreatureType("Beast")),
				TargetCreature("target creature an opponent controls", OpponentControls()),
			),
			Effect: slotZeroFightsSlotOne,
		}},
	})
}
