package effects

// Dragon Sniper — Creature — Human Archer {G}, 1/1 (EDHREC rank
// 4062):
//
//	"Reach, vigilance, deathtouch"
//
// A one-mana vanilla with three keywords and no text, which is
// exactly why it is worth a slot: the three keywords compose into a
// board presence far larger than a 1/1. Reach plus deathtouch is a
// wall that eats any flier that attacks into it; vigilance means the
// Sniper is still standing there after it has attacked. In a
// Commander pod that reads as "nobody's Dragon attacks this player"
// for one green mana.
//
// Every keyword rides PrintedKeywords and has a live consumer in the
// engine: reach and deathtouch are read by the block-legality and
// lethal-damage checks, vigilance skips the attack tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "913dbce2-6805-453f-a9c3-64f1cc658627",
		Name:            "Dragon Sniper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "deathtouch"},
	})
}
