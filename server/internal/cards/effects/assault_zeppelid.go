package effects

// Assault Zeppelid — Creature — Beast {2}{G}{U}, 3/3 (Dissension, #5598):
//
//	"Flying, trample"
//
// Two printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8049dbff-6b57-454a-a2e6-7e2ae39bb606",
		Name:            "Assault Zeppelid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
}
