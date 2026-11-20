package effects

// Sovereign's Bite — Sorcery {1}{B} (Core Set 2019, #5527):
//
//	"Target player loses 3 life and you gain 3 life."
//
// Soul Feast's text with 3 life: the shared
// targetPlayerLosesLifeYouGain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13718d93-fe46-4a18-a511-88649ef1c9de",
		Name:         "Sovereign's Bite",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerLosesLifeYouGain(3),
	})
}
