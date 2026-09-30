package effects

// Capture of Jingzhou — Sorcery {3}{U}{U}:
//
//	"Take an extra turn after this one."
//
// Temporal Manipulation's text under another name; see that file.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89aa65d9-2502-40b0-90b6-b25a8e9f6155",
		Name:         "Capture of Jingzhou",
		Completeness: CompletenessFull,
		OnResolve:    youTakeAnExtraTurn,
	})
}
