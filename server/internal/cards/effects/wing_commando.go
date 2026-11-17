package effects

// Wing Commando — Creature — Human Soldier {2}{U}, 2/2 (The Brothers' War, #5482):
//
//	"Flying
//	 Prowess"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "423c8465-82cd-4f54-b5ac-022b7ebcf3e4",
		Name:            "Wing Commando",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
