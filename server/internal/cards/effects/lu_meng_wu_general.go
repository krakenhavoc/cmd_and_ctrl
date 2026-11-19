package effects

// Lu Meng, Wu General — Legendary Creature — Human Soldier {3}{U}{U}, 4/4
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
		OracleID:        "a1e2a966-6aad-4325-b644-3ced319bdd7d",
		Name:            "Lu Meng, Wu General",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
