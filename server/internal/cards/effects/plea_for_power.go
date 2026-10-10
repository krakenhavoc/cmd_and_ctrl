package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plea for Power — Sorcery {3}{U} (EDHREC rank 5655):
//
//	"Will of the council — Starting with you, each player votes for
//	 time or knowledge. If time gets more votes, take an extra turn
//	 after this one. If knowledge gets more votes or the vote is tied,
//	 draw three cards."
//
// The vote is ADR 0146's (CR 701.38): each player in turn order from
// the caster, with any extra votes, each one logged as it is cast. A
// tie goes to knowledge, as printed.
//
// A bot casting it votes time; an opponent's bot votes knowledge, the
// smaller of the two gifts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b886623-5c96-48dd-a3d2-fd5865c58ff5",
		Name:         "Plea for Power",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Plea for Power — vote for time or knowledge",
				Words:         []string{"time", "knowledge"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          pleaForPowerVoted,
			}.Apply(ctx)
		},
	})
}

var pleaForPowerVoted = VoteResultThen("vote/plea-for-power", func(ctx *Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		return TakeExtraTurn{}.Apply(ctx)
	}
	return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
})
