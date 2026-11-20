package effects

// Grave Bramble — Creature — Plant {1}{G}{G}, 3/4 (Innistrad, #5572):
//
//	"Defender, protection from Zombies"
//
// Defender and protection from a creature type are engine keywords
// carried as printed data: it can't attack (CR 702.3b), and a Zombie
// can't block it, be blocked by it, damage it, enchant or equip it, or
// target it (CR 702.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e2f854b-17f7-4d52-927c-abdfef86c1f8",
		Name:            "Grave Bramble",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "protection from Zombies"},
	})
}
