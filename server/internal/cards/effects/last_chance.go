package effects

// Last Chance — Sorcery {R}{R}:
//
//	"Take an extra turn after this one. At the beginning of that turn's
//	 end step, you lose the game."
//
// Final Fortune at sorcery speed; see final_fortune.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "360039a5-1cbd-4ee3-8f94-21b5348e106a",
		Name:         "Last Chance",
		Completeness: CompletenessFull,
		OnResolve:    extraTurnThenLoseAtItsEndStep("Last Chance — you lose the game"),
	})
}
