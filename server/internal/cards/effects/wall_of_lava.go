package effects

// Wall of Lava — Creature — Wall {1}{R}{R}, 1/3 (Ice Age, #5636):
//
//	"Defender
//	 {R}: This creature gets +1/+1 until end of turn."
//
// Defender is a printed keyword (CR 702.3b: it can't attack, and it
// blocks as usual). The pump is the shared selfPumpRow at +1/+1 for
// {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bf7734fb-011d-4f12-a174-1ae879dab95b",
		Name:            "Wall of Lava",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []ActivatedAbility{selfPumpRow("Wall of Lava", "{R}", 1, 1)},
	})
}
