package effects

// Disciplined Duelist — Creature — Human Citizen {G}{W}{U}, 2/1
// (Streets of New Capenna):
//
//	"Double strike
//	 This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c650f497-92d6-48cf-8b59-5d61eb184dc0",
		Name:            "Disciplined Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Replacements:    snc6Entry("Disciplined Duelist"),
	})
}
