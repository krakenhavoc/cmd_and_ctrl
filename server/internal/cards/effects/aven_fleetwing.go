package effects

// Aven Fleetwing — Creature — Bird Soldier {3}{U}, 2/2 (Magic 2012, #5573):
//
//	"Flying
//	 Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// Flying and hexproof, both printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37ad9318-6690-4a8d-bcea-9fcbca58769d",
		Name:            "Aven Fleetwing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
	})
}
