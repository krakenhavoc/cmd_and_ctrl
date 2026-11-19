package effects

// Greenwood Sentinel — Creature — Elf Scout {1}{G}, 2/2 (Foundations
// Jumpstart):
//
//	"Vigilance"
//
// A printed keyword. No simplification.
func init() {
	Register(Spec{
		OracleID:        "8178c546-84f8-4563-a804-6489ec016660",
		Name:            "Greenwood Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
