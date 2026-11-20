package effects

// Undercity Shade — Creature — Shade {4}{B}, 1/1 (Ravnica: City of Guilds,
// #5600):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Fear is printed data (CR 702.36); the pump is Killer Bees' legSelfPump,
// any number of activations, each ending at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28f3d535-9503-4d35-aeed-d3e59259f760",
		Name:            "Undercity Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated:       []ActivatedAbility{legSelfPump("{B}", 1, "This creature", "Undercity Shade")},
	})
}
