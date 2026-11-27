package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pain's Reward — Sorcery {2}{B} (Saviors of Kamigawa, #5601):
//
//	"Each player may bid life. You start the bidding with a bid of any
//	 number. In turn order, each player may top the high bid. The bidding
//	 ends if the high bid stands. The high bidder loses life equal to the
//	 high bid and draws four cards."
//
// LifeBidding (#6212) runs the auction over every player still in the game,
// asking the caster for the opening bid (any number, 0 up) and then each
// player in turn order, and hands the high bid and the high bidder to the
// win. The win is a registered number branch: the bidder loses that much life
// (CR 119.3) and draws four cards. A bid is only a number (CR 107.1b), so it
// may be more life than the bidder has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4608c663-e088-40ca-a915-b6e39253eb98",
		Name:         "Pain's Reward",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.LifeBidding{
			OpeningChosen: true,
			Then:          effects.OldB585PainsRewardWon,
		}),
	})
}
