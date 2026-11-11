package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wavesifter — Creature — Elemental {3}{G}{U}, 3/2 (Murders at Karlov
// Manor Commander, #5368):
//
//	"Flying
//	 When this creature enters, investigate twice.
//	 Evoke {G}{U}"
//
// Mulldrifter's shape: the enters trigger is real and so is the evoke
// sacrifice (CR 702.74a), so an evoked Wavesifter still makes its two
// Clues before it is sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d1ec1e4e-faa6-4afc-9e00-a3327105c11b",
		Name:             "Wavesifter",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Evoke("{G}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wavesifter — investigate twice", Do(Investigate{N: 2})),
		},
	})
}
