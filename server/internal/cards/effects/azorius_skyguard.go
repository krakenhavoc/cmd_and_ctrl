package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Skyguard — Creature — Human Knight {4}{W}{U}, 3/3 (Ravnica
// Allegiance, #5521):
//
//	"Flying, first strike
//	 Creatures your opponents control get -1/-0."
//
// Flying and first strike are printed data (CR 702.9b, 702.7b); the -1/-0
// is a layer-7c static over every creature an opponent controls, Waker of
// Waves's line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bbfdb5eb-2140-4d50-a667-1b4ac75d03a5",
		Name:            "Azorius Skyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Static:          []game.StaticAbility{b16Anthem(b27CreaturesOpponentsControl, -1, 0)},
	})
}
