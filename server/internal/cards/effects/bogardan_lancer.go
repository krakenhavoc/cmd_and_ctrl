package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bogardan Lancer — Creature — Human Knight {1}{R}, 1/1 (Future Sight, #5594):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Flanking (Whenever a creature without flanking blocks this
//	 creature, the blocking creature gets -1/-1 until end of turn.)"
//
// Bloodthirst is the shared CR 702.54a replacement (Burning-Tree
// Bloodscale) and flanking the engine's printed-keyword trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67953edb-9681-4339-a5d2-917082925c24",
		Name:            "Bogardan Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flanking"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(1, "Bogardan Lancer")},
	})
}
