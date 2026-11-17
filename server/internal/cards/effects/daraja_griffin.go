package effects

// Daraja Griffin — Creature — Griffin {3}{W}, 2/2 (Visions, #5631):
//
//	"Flying
//	 Sacrifice this creature: Destroy target black creature."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// sacrifice is a cost, so the Griffin is gone before the ability
// resolves; a target that is no longer black, or has left, is an illegal
// target and nothing is destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "daa0d720-cac1-487a-93e2-33e18f91f19b",
		Name:            "Daraja Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target black creature.",
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target black creature", OfColor("B")),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
