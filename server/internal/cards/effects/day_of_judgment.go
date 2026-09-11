package effects

// Day of Judgment — "Destroy all creatures." Wrath of God without
// the "can't be regenerated" clause, which is a distinction with no
// difference until regeneration is modelled. Shares the
// wrathDestroyAllCreatures helper defined in wrath_of_god.go.
func init() {
	Register(Spec{
		OracleID:  "d057289d-5e28-43d5-8ff3-4a1bc723477d",
		Name:      "Day of Judgment",
		OnResolve: wrathDestroyAllCreatures,
	})
}
