package effects

// Bird Maiden — Creature — Human Bird {2}{R}, 1/2 (Arabian Nights, #5641):
//
//	"Flying"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bddd6b93-7c69-427a-97ee-0c1f9f950f09",
		Name:            "Bird Maiden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
