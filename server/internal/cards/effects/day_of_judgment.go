package effects

// Day of Judgment — "Destroy all creatures." Mass-wipe without the
// "can't be regenerated" clause (cosmetic in S14 — no regeneration
// pipeline until S17 / S18). Shares the wrathDestroyAllCreatures
// helper defined in wrath_of_god.go.
func init() {
	Register(Spec{
		OracleID:  "d057289d-5e28-43d5-8ff3-4a1bc723477d",
		Name:      "Day of Judgment",
		OnResolve: wrathDestroyAllCreatures,
	})
}
