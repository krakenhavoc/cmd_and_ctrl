package effects

// Healer's Flock — Creature — Bird {W}{W}{W}, 3/3 (Modern Horizons 2,
// #5501):
//
//	"Flying, lifelink"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0b3b2fb1-d5fe-4d14-985a-85f13cd60d9c",
		Name:            "Healer's Flock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
