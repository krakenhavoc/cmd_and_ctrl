package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungry Graffalon — Creature — Giraffe {3}{G}, 3/4 (Secrets of Strixhaven,
// #4456):
//
//	"Reach
//	 Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)"
//
// Reach is a printed keyword; Increment is the keyword helper (CR 702.191),
// which reads the mana spent as the spell is cast and rechecks it as the
// ability resolves (Cuboid Colony's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9494c499-6ead-491f-aeb0-6dbfac4d960c",
		Name:            "Hungry Graffalon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{Increment()},
	})
}
