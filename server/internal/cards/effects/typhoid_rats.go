package effects

// Typhoid Rats — "Deathtouch."
//
// Minimal deathtouch creature — 1/1 for B. Combat hits flag the
// target via MarkedLethalByDeathtouch so the SBA destroys
// regardless of remaining toughness. Exit criterion for S18: a
// 1/1 deathtouch attacker takes down a 5/5 blocker.
func init() {
	Register(Spec{
		OracleID:        "d6ee6cc1-902d-4f56-afa5-6fa4813bfbbc",
		Name:            "Typhoid Rats",
		PrintedKeywords: []string{"deathtouch"},
	})
}
