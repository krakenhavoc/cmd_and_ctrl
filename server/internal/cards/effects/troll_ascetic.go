package effects

// Troll Ascetic — Creature — Troll Shaman {1}{G}{G}, 3/2 (Mirrodin, #5606):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)
//	 {1}{G}: Regenerate this creature."
//
// Printed hexproof (CR 702.11b) and the shared regenerate row
// (regenerate_this.go): a shield on the ability's source, used the
// next time it would be destroyed this turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0585ea6a-ef78-4386-a22c-6b8d8fc6c03c",
		Name:            "Troll Ascetic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Activated:       []ActivatedAbility{regenerateThisRow("{1}{G}")},
	})
}
