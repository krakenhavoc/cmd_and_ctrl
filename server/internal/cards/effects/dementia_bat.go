package effects

// Dementia Bat — Creature — Phyrexian Bat {4}{B}, 2/2 (New Phyrexia, #5575):
//
//	"Flying
//	 {4}{B}, Sacrifice this creature: Target player discards two cards."
//
// Flying, and a sacrifice row whose target player is asked to discard
// two (targetOpponentDiscards reads any target player). A player with
// fewer cards discards what they have (CR 701.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95c76175-e3a6-4bf4-a8da-d1179851efe5",
		Name:            "Dementia Bat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}, Sacrifice this creature: Target player discards two cards.",
			Cost:    Plus(ManaCost("{4}{B}"), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  targetOpponentDiscards(2),
		}},
	})
}
