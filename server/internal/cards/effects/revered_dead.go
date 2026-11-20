package effects

// Revered Dead — Creature — Spirit Soldier {1}{W}, 1/1 (Planar Chaos, #5595):
//
//	"{W}: Regenerate this creature."
//
// Regeneration on the ability's own source (CR 701.19), lrwB1RegenerateThisRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96a57ce0-f1ab-4af7-8c18-8e35aad11780",
		Name:         "Revered Dead",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{lrwB1RegenerateThisRow("{W}")},
	})
}
