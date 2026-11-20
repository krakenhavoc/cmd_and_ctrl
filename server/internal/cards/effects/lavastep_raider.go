package effects

// Lavastep Raider — Creature — Goblin Warrior {R}, 1/2 (Battle for
// Zendikar, #5547):
//
//	"{2}{R}: This creature gets +2/+0 until end of turn."
//
// The firebreathing row (thisGetsRow): its cost has no {T}, so it can
// be activated any number of times, the turn it arrives included; each
// pump ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a96782b-e9c8-4bc5-845d-45f6bf7b8dc4",
		Name:         "Lavastep Raider",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{2}{R}", 2, 0, "Lavastep Raider")},
	})
}
