package effects

// Elite Inquisitor — Creature — Human Soldier {W}{W}, 2/2 (Innistrad, #5572):
//
//	"First strike, vigilance
//	 Protection from Vampires, from Werewolves, and from Zombies"
//
// Engine keywords carried as printed data. "Protection from Vampires,
// from Werewolves, and from Zombies" is three protection abilities
// (CR 702.16m), each a creature-type quality the protection parser reads
// (Katilda, Dawnhart Martyr's "protection from Vampires").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f7e865b-6ab0-4361-b864-63c38cdb2fed",
		Name:            "Elite Inquisitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "protection from Vampires", "protection from Werewolves", "protection from Zombies"},
	})
}
