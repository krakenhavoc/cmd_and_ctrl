package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ascending Aven — Creature — Bird Soldier {2}{U}{U}, 3/2 (Onslaught):
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 Morph {2}{U}"
//
// Brazen Borrower's block rule (CR 509.1b) on a morph. Face down it is a
// 2/2 with no text (CR 708.2), so it has neither flying nor the block
// rule and may block anything; turned face up for {2}{U} it has both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "67b4a4bc-0b02-4cef-b5c8-41c839eaf0cd",
		Name:             "Ascending Aven",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{2}{U}")},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
