package effects

// Wall of Faith — Creature — Wall {3}{W}, 0/5 (Magic 2010, #5585):
//
//	"Defender (This creature can't attack.)
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Defender rides PrintedKeywords; the pump is the shared self-pump row,
// a layer 7c change that ends at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "918d6cd4-76d5-4e21-b92b-1ecf35eb6cff",
		Name:            "Wall of Faith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []ActivatedAbility{selfPumpRow("Wall of Faith", "{W}", 0, 1)},
	})
}
