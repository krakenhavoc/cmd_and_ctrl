package effects

// Towering Viewpoint — Artifact Creature — Wall {2}, 0/4 (Assassin's
// Creed, #5466):
//
//	"Defender, reach
//	 Leap of Faith — {3}: Target creature gains flying until end of
//	 turn."
//
// "Leap of Faith" is an ability word with no rules meaning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46c404c3-636a-4945-a61c-62f013a9c50f",
		Name:            "Towering Viewpoint",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []ActivatedAbility{{
			Label:   "Leap of Faith — {3}: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    ManaCost("{3}"),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Towering Viewpoint — flying until end of turn"),
		}},
	})
}
