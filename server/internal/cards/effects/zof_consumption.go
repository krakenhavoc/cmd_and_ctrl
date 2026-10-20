package effects

// Zof Consumption // Zof Bloodbog — Modal DFC. Front: Sorcery {4}{B}{B}:
// "Each opponent loses 4 life and you gain 4 life." Back: Land
// (mdfc_lands.go).
//
// Two steps (CR 608.2c): the life loss, which can stop on a CR 616 ordering
// prompt, then the gain, so the gain waits for the answer. No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "d9f11985-e460-425d-b083-9cb0edf1983a",
		Name:         "Zof Consumption",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(EachOpponentLosesLife{N: 4}, GainLife{Amount: 4}),
	})
}
