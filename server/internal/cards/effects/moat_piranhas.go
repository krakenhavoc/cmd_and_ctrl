package effects

// Moat Piranhas — Creature — Fish {1}{U}, 3/3 (Core Set 2020, #5518):
//
//	"Defender (This creature can't attack.)"
//
// A printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40057917-4a88-4316-bc5b-efc0ada1ddff",
		Name:            "Moat Piranhas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
