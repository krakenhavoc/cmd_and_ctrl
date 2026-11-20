package effects

// Return to the Earth — Instant {3}{G} (Fate Reforged,
// #5552):
//
//	"Destroy target artifact, enchantment, or creature with flying."
//
// Broken Wings' text and body: the creature clause is a creature with flying
// as it is now, re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "722f2965-a1ec-4b97-a4bd-b99364f58b1e",
		Name:         "Return to the Earth",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: destroyTheTargetPermanent,
	})
}
