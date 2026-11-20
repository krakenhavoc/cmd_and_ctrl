package effects

// Wei Elite Companions — Creature — Human Soldier {4}{B}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2db5e0cd-c991-4fda-9001-9e4f10e8a805",
		Name:            "Wei Elite Companions",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
