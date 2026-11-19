package effects

// Aven Reedstalker — Creature — Bird Warrior {3}{U}, 2/3 (Foundations
// Jumpstart, #5929):
//
//	"Flash
//	 Flying"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d9f43942-4296-417d-a180-5b4f64e4b12d",
		Name:            "Aven Reedstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
