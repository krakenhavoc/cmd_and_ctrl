package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exalted Sunborn — Creature — Angel Wizard {3}{W}{W}, 4/5 (Edge of
// Eternities, #4461):
//
//	"Flying, lifelink
//	 If one or more tokens would be created under your control, twice
//	 that many of those tokens are created instead.
//	 Warp {1}{W}"
//
// Mondrak, Glory Dominus's token doubler (TokensDoubled, the CR 614
// replacement Parallel Lives was built on) in a flying, lifelink body
// with warp (CR 702.185). A warped Sunborn doubles tokens until it is
// exiled at the next end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "43febbcd-ecbe-4edb-ad90-7fc6dc381d05",
		Name:             "Exalted Sunborn",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{W}")},
		Replacements: []game.ReplacementEffect{
			effects.TokensDoubled("Exalted Sunborn: double tokens"),
		},
	})
}
