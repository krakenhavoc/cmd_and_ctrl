package effects

// Broken Wings — Instant {2}{G} (Aetherdrift, #4464):
//
//	"Destroy target artifact, enchantment, or creature with flying."
//
// Make Your Move's destroy shape with the creature clause narrowed to a
// creature with flying, read off the permanent as it is now (a creature
// that only gained flying until end of turn is a legal target). The
// target is re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e316864-d55c-496f-8f46-773567896864",
		Name:         "Broken Wings",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: destroyTheTargetPermanent,
	})
}
