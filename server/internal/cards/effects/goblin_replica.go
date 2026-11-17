package effects

// Goblin Replica — Artifact Creature — Goblin {3}, 2/2 (Mirrodin, #5606):
//
//	"{3}{R}, Sacrifice this creature: Destroy target artifact."
//
// Elf Replica's shape aimed at an artifact: the sacrifice is paid at
// announce, the target is checked again as the ability resolves
// (CR 608.2b), and an indestructible artifact survives. "Target
// artifact" admits the Replica itself as it is announced, but it is
// sacrificed as the cost, so it is gone before the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c5a7772-8358-4549-a583-f54246869e20",
		Name:         "Goblin Replica",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice this creature: Destroy target artifact.",
			Cost:    Plus(ManaCost("{3}{R}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
