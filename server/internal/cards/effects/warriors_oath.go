package effects

// Warrior's Oath — Sorcery {R}{R}:
//
//	"Take an extra turn after this one. At the beginning of that turn's
//	 end step, you lose the game."
//
// Last Chance's text under another name; see final_fortune.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "574044cf-2e2f-4b5c-b4d1-cf05ba814ab2",
		Name:         "Warrior's Oath",
		Completeness: CompletenessFull,
		OnResolve:    extraTurnThenLoseAtItsEndStep("Warrior's Oath — you lose the game"),
	})
}
