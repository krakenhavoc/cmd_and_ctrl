package effects

// Azorius First-Wing — Creature — Griffin {W}{U}, 2/2 (Dissension, #5598):
//
//	"Flying, protection from enchantments"
//
// Flying and one protection whose quality is a card type (CR 702.16a):
// an enchantment can't enchant it, target it, block it or damage it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "08d02fdb-4b3f-4f55-b5fc-6bb4d2a14b1f",
		Name:            "Azorius First-Wing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from enchantments"},
	})
}
