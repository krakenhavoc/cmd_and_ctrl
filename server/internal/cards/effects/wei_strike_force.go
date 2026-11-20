package effects

// Wei Strike Force — Creature — Human Soldier {2}{B}, 2/1
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
		OracleID:        "002ad5fa-a20c-4d30-beef-401d89868d67",
		Name:            "Wei Strike Force",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
