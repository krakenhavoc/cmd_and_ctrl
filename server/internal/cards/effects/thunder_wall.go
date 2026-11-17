package effects

// Thunder Wall — Creature — Wall {1}{U}{U}, 0/2 (Ice Age, #5636):
//
//	"Defender
//	 Flying
//	 {U}: This creature gets +1/+1 until end of turn."
//
// Defender and flying are printed keywords, so it can't attack
// (CR 702.3b) and can block fliers. The pump is the shared selfPumpRow
// at +1/+1 for {U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "904b9056-ad76-4805-b14f-c426e4a57bae",
		Name:            "Thunder Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated:       []ActivatedAbility{selfPumpRow("Thunder Wall", "{U}", 1, 1)},
	})
}
