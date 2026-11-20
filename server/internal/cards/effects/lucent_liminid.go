package effects

// Lucent Liminid — Enchantment Creature — Elemental {3}{W}{W}, 3/3
// (Future Sight, #5594):
//
//	"Flying"
//
// A printed keyword and nothing else; the enchantment type rides the
// type line from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf02b774-2b3a-4459-bf59-e5d291691e77",
		Name:            "Lucent Liminid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
