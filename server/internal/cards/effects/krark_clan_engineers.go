package effects

// Krark-Clan Engineers — Creature — Goblin Artificer {3}{R}, 2/2 (Fifth Dawn, #5604):
//
//	"{R}, Sacrifice two artifacts: Destroy target artifact."
//
// Goblin Replica's artifact kill, paid with {R} and Metalwork Colossus's
// "sacrifice two artifacts" (one simultaneous exit).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0964288c-0397-428a-9a84-784aaea4e040",
		Name:         "Krark-Clan Engineers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice two artifacts: Destroy target artifact.",
			Cost:    Plus(ManaCost("{R}"), SacrificeN(2, "two artifacts", Artifact())),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
