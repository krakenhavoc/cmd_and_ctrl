package effects

// Wall of Torches — Creature — Wall {1}{R}, 4/1 (Magic 2012, #5573):
//
//	"Defender (This creature can't attack.)"
//
// Defender is a printed keyword (Glacial Wall's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "674a269c-61ed-4fb5-ad74-422e581a5a91",
		Name:            "Wall of Torches",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
