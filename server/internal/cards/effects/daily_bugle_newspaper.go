package effects

// Daily Bugle Newspaper — Artifact {2} (Marvel Super Heroes, #4455):
//
//	"{2}, {T}: Draw a card, then discard a card. Create a Treasure token."
//
// Collector's Vault under a Marvel name: the ability is
// lootAndTreasureAbility, the loot first (the discard is the controller's
// choice and the Treasure waits for it, CR 608.2c) and the Treasure after.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2a3f735-9523-4ddd-96a9-b0a5ecd39284",
		Name:         "Daily Bugle Newspaper",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{lootAndTreasureAbility()},
	})
}
