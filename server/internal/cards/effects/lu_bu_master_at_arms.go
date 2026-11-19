package effects

// Lu Bu, Master-at-Arms — Legendary Creature — Human Soldier Warrior {5}{R}, 4/3
// (Portal Three Kingdoms, #5622):
//
//	"Haste; horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff959f1f-0982-4761-a6b0-3044dff1d617",
		Name:            "Lu Bu, Master-at-Arms",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "horsemanship"},
	})
}
