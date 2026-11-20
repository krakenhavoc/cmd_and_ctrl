package effects

// Manifold Key — Artifact {1} (Core Set 2020, #5518):
//
//	"{1}, {T}: Untap another target artifact.
//	 {3}, {T}: Target creature can't be blocked this turn."
//
// Sonic Screwdriver's untap and can't-be-blocked abilities, sharing one
// {T}. "Another" is object identity (CR 109.1): a second Key can be
// untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "370c37f7-2484-40e8-8fd9-c10cf6f5d5f2",
		Name:         "Manifold Key",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap another target artifact.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: Another(TargetPermanent("another target artifact", Artifact())),
				Effect:  untapFirstLegalTarget,
			},
			{
				Label:   "{3}, {T}: Target creature can't be blocked this turn.",
				Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  targetCantBeBlockedThisTurn("Manifold Key"),
			},
		},
	})
}
