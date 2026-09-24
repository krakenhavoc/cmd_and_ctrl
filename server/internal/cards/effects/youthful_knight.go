package effects

// Youthful Knight — "First strike."
//
// Minimal first-strike creature — 2/1 for 1W. Participates in the
// CR 510.4 first-strike substep only; survives trades with
// 2-toughness vanilla blockers because the blocker dies before
// regular-damage assignment.
func init() {
	Register(Spec{
		OracleID:        "ef2a24f5-ce5e-4054-843a-2cae0c66318a",
		Name:            "Youthful Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
