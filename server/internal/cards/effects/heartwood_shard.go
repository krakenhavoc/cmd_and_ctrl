package effects

// Heartwood Shard — Artifact {3} (Mirrodin, #5606):
//
//	"{3}, {T} or {G}, {T}: Target creature gains trample until end of
//	 turn."
//
// Crystal Shard's two-row shape (one ability, a choice of cost, picked by
// picking the row). Both rows give the target trample until end of turn
// as a layer-6 grant pinned to it (ebTargetCreatureGainsUntilEOT); a
// target that is no longer legal as the ability resolves gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	const effect = "Target creature gains trample until end of turn."
	grant := ebTargetCreatureGainsUntilEOT("trample", "Heartwood Shard — trample until end of turn")
	Register(Spec{
		OracleID:     "64d4bdc5-ff40-469d-bb25-82de0a49dfa4",
		Name:         "Heartwood Shard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: " + effect,
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "trample")),
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  grant,
			},
			{
				Label:   "{G}, {T}: " + effect,
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "trample")),
				Cost:    Plus(ManaCost("{G}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  grant,
			},
		},
	})
}
