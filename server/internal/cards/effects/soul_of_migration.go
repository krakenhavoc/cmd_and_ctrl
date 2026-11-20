package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul of Migration — Creature — Elemental {5}{W}{W}, 2/4 (Modern
// Horizons 2, #5501):
//
//	"Flying
//	 When this creature enters, create two 1/1 white Bird creature tokens
//	 with flying.
//	 Evoke {3}{W} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Shriekmaw's evoke (CR 702.74a): the enter trigger and the evoke
// sacrifice trigger both go on the stack, and the controller orders
// them, so the Birds arrive either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f90ebf35-08f9-4277-92d9-75165f32a84e",
		Name:             "Soul of Migration",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		AlternativeCosts: []game.AlternativeCost{Evoke("{3}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Soul of Migration — create two 1/1 white Bird creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 2})),
		},
	})
}
