package effects

// Mistmeadow Skulk — Creature {1}{W}, 1/1:
//
//	"Lifelink, protection from mana value 3 or greater"
//
// "Or greater" is game.ProtectionQualityManaValueAtLeast, the mirror of
// Reaver Titan's "or less".
func init() {
	Register(Spec{
		OracleID:        "490df3b5-ece2-45bb-af38-fcd7f83e37b0",
		Name:            "Mistmeadow Skulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "protection from mana value 3 or greater"},
	})
}
