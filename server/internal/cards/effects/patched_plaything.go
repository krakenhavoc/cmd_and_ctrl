package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patched Plaything — Artifact Creature — Toy {2}{W}, 4/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Double strike
//	 This creature enters with two -1/-1 counters on it if you cast it
//	 from your hand."
//
// The counters are a CR 614.1c entry clause read off the resolving
// spell's cast record (CountersIfCastFrom, Storied Veteran's): cast from
// your hand it enters as a 2/1 double striker. Cast from anywhere else,
// or put onto the battlefield without being cast (reanimated, blinked),
// it enters as the printed 4/3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0b7e950a-89e1-4ec4-94af-5af2f1252b12",
		Name:            "Patched Plaything",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfCastFrom(game.ZoneHand, game.CounterMinusOne, 2),
		},
	})
}
