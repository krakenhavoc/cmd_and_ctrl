package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyrant's Choice — Sorcery {1}{B}:
//
//	"Will of the council — Starting with you, each player votes for
//	 death or torture. If death gets more votes, each opponent
//	 sacrifices a creature of their choice. If torture gets more votes
//	 or the vote is tied, each opponent loses 4 life."
//
// The vote is ADR 0146's (CR 701.38); a tie goes to torture, as
// printed. The edict is the ordinary each-opponent sacrifice, one
// prompt per opponent addressed to that opponent.
//
// A bot casting it votes torture; an opponent's bot votes death, since
// a spare creature costs less than four life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee7d2c2f-e32f-4be6-82ac-ff1e32055376",
		Name:         "Tyrant's Choice",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Tyrant's Choice — vote for death or torture",
				Words:         []string{"death", "torture"},
				ForController: []int{1, 2},
				ForOpponents:  []int{1, 0},
				Then:          tyrantsChoiceVoted,
			}.Apply(ctx)
		},
	})
}

var tyrantsChoiceVoted = VoteResultThen("vote/tyrants-choice", func(ctx *Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		return EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}.Apply(ctx)
	}
	return EachOpponentLosesLife{N: 4}.Apply(ctx)
})
