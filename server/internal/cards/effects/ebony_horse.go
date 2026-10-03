package effects

// Ebony Horse — Artifact {3}:
//
//	"{2}, {T}: Untap target attacking creature you control. Prevent all
//	 combat damage that would be dealt to and dealt by that creature this
//	 turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Maze of Ith's ability on your own
// attacker — one to-and-by record (Mod.AndDealtBy).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "4eb67ba3-35e3-47c3-820f-62814cf202a7",
		Name:         "Ebony Horse",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{untapAttackerToAndByRow(
			"{2}, {T}: Untap target attacking creature you control. Prevent all combat damage that would be dealt to and dealt by that creature this turn.",
			Plus(ManaCost("{2}"), TapCost()),
			TargetCreature("target attacking creature you control", AttackingCreature(), YouControl()))},
	})
}
