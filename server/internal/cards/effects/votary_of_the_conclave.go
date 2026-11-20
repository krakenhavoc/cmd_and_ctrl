package effects

// Votary of the Conclave — Creature — Human Soldier {W}, 1/1 (Ravnica: City
// of Guilds, #5600):
//
//	"{2}{G}: Regenerate this creature."
//
// The regeneration row is regenerateThisRow (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83837ceb-7aae-4257-8545-a6cf34656e3b",
		Name:         "Votary of the Conclave",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{regenerateThisRow("{2}{G}")},
	})
}
