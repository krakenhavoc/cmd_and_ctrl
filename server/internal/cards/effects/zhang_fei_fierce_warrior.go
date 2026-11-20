package effects

// Zhang Fei, Fierce Warrior — Legendary Creature — Human Soldier Warrior {4}{W}{W}, 4/4
// (Portal Three Kingdoms, #5622):
//
//	"Vigilance; horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3f815b8-75c6-40e0-9716-872f0f5c9d46",
		Name:            "Zhang Fei, Fierce Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "horsemanship"},
	})
}
