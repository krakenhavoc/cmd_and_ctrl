package effects

// Magewright's Stone — Artifact {2} (EDHREC rank 2890):
//
//	"{1}, {T}: Untap target creature that has an activated ability
//	 with {T} in its cost."
//
// The untapper for tap-ability creatures. The target clause is
// b27HasTapAbility: a creature counts when the engine would offer
// its controller an activated ability or a mana ability with a tap
// in the cost — a catalog creature's declared abilities, a token's
// intrinsic ones, a land creature's intrinsic mana. That is exactly
// the set of creatures the Stone could usefully untap here, and it
// is narrower than the printed clause for one reason only: a
// creature whose tap ability is not in the catalog has no ability
// the engine can see, and so cannot be targeted.
//
// Sandbox simplification, declared: creatures the catalog does not
// know are not legal targets, even if their printed text has a {T}
// ability. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "9e500094-c430-4c10-89cc-97f4f18b08bc",
		Name:         "Magewright's Stone",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only creatures whose tap abilities the engine automates can be targeted."},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target creature that has an activated ability with {T} in its cost.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature that has an activated ability with {T} in its cost", b27HasTapAbility()),
			Effect:  untapTheTarget,
		}},
	})
}
