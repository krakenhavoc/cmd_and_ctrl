package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slippery Scoundrel — Creature — Human Pirate {2}{U}, 2/2 (EDHREC
// rank 16357):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 As long as you have the city's blessing, this creature has hexproof
//	 and can't be blocked."
//
// Both halves are gated on the controller's blessing, read live: the
// hexproof is a layer-6 keyword grant (the layer pass is invalidated
// when the blessing is earned), the evasion a CR 509.1b block rule on
// the creature itself. The blessing is never lost once earned (CR
// 702.131c), so a Scoundrel whose controller's board later shrinks
// stays hexproof and unblockable.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c7ea8d0-d866-4c02-bdfc-13bbc87c094a",
		Name:            "Slippery Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Static:          []game.StaticAbility{SelfKeywordWhileCitysBlessing("hexproof")},
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(OnSelf(), youHaveTheCitysBlessingCast),
		},
	})
}
