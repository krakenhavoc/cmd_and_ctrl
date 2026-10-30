package effects

// Make Your Move — Instant {2}{W} (Marvel Super Heroes Commander,
// #4455): "Destroy target artifact, enchantment, or creature with
// power 4 or greater." Abolish's destroy shape, with the target's
// predicate widened to a creature with power 4 or greater.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8226f31d-6f51-49c3-87f7-0c68f7f4f9ce",
		Name:         "Make Your Move",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with power 4 or greater",
			Or(Artifact(), Enchantment(), And(Creature(), PowerGE(4)))),
		OnResolve: destroyTheTargetPermanent,
	})
}
