package effects

// Wall of Stone — "Defender."
//
// Canonical defender wall — 0/8 for RR. DeclareAttacker rejects
// with ErrDefender so this creature cannot attack, but CanBlock
// still accepts it (defender only restricts attacking, not
// blocking, per CR 702.3b).
func init() {
	Register(Spec{
		OracleID:        "cd4cadb4-3156-49bd-b36e-12ba5c85938b",
		Name:            "Wall of Stone",
		PrintedKeywords: []string{"defender"},
	})
}
