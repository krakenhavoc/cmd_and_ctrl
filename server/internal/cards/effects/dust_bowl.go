package effects

// Dust Bowl — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice a land: Destroy target nonbasic land."
//
// Wasteland with the sacrifice moved off the source: any land you
// control pays it, this one included. The sacrifice is a cost, so it is
// gone before the ability resolves; the target is chosen at announce and
// can't be the land sacrificed for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d3df7128-31dd-4d71-90be-87e2e9ff51b4",
		Name:          "Dust Bowl",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice a land: Destroy target nonbasic land.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), b08SacrificeALand()),
			Targets: TargetPermanent("target nonbasic land", Land(), b10NonbasicLand()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
