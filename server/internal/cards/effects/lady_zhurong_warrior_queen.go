package effects

// Lady Zhurong, Warrior Queen — Legendary Creature — Human Soldier Warrior {4}{G}, 4/3
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
		OracleID:        "531e6ffb-28e0-4e8e-ad88-2f3a243789cf",
		Name:            "Lady Zhurong, Warrior Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
