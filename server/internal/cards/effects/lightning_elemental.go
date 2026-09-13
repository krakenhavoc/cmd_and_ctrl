package effects

// Lightning Elemental — "Haste."
//
// Canonical haste vanilla — 4/1 for 3R. HasSummoningSickness
// returns false because of the read-time haste bypass, so
// DeclareAttacker accepts the creature the same turn it ETBs.
func init() {
	Register(Spec{
		OracleID:        "58aee5cb-7b88-446e-ab10-9f83c10d7227",
		Name:            "Lightning Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
