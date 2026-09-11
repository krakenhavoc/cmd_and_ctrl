package effects

// Damnation — "Destroy all creatures. They can't be regenerated."
// Functional mirror of Wrath of God in a different colour. Shares
// the wrathDestroyAllCreatures helper defined in wrath_of_god.go,
// which since S23 sweeps through DestroyAllMatching so the deaths
// are simultaneous.
func init() {
	Register(Spec{
		OracleID:  "d57a8f0b-7989-4db5-8756-6f2690097252",
		Name:      "Damnation",
		OnResolve: wrathDestroyAllCreatures,
	})
}
