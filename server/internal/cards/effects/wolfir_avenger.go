package effects

// Wolfir Avenger — {1}{G}{G} Creature — Wolf Warrior 3/3 (Avacyn Restored, #5570):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 {1}{G}: Regenerate this creature."
//
// Flash and the shared regeneration row (regenerate_this.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83deda94-66e8-4655-9cd5-6d97d809946d",
		Name:            "Wolfir Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []ActivatedAbility{regenerateThisRow("{1}{G}")},
	})
}
