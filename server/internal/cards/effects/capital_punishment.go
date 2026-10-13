package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Capital Punishment — Sorcery {4}{B}{B}:
//
//	"Council's dilemma — Starting with you, each player votes for
//	 death or taxes. Each opponent sacrifices a creature of their choice
//	 for each death vote and discards a card for each taxes vote."
//
// A council's dilemma (ADR 0146, CR 701.38): every ballot counts. Each
// opponent sacrifices as many creatures as there are death votes, each
// chosen by that opponent, and then discards as many cards as there
// are taxes votes, also their choice. The sacrifices come first because
// the sentence prints them first, so a sacrifice's dies-triggers
// resolve before anyone picks a discard. An opponent with fewer
// creatures than death votes sacrifices what they have (CR 701.21a).
//
// A bot casting it votes death; an opponent's bot votes taxes, since
// a card from hand costs less than a creature from the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2902f4f-b40e-49f7-8ab0-5b6a70021482",
		Name:         "Capital Punishment",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Capital Punishment — vote for death or taxes",
				Words:         []string{"death", "taxes"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          capitalPunishmentVoted,
			}.Apply(ctx)
		},
	})
}

var capitalPunishmentVoted = VoteResultThen("vote/capital-punishment", func(ctx *Context, r game.VoteResult) error {
	death, taxes := r.Votes(0), r.Votes(1)
	discard := func(g *game.Game, _ game.PromptedSacrifices) error {
		eachOpponentDiscards(NewContext(g, ctx.Item), taxes)
		return nil
	}
	var seats []uuid.UUID
	for _, opp := range ctx.Opponents() {
		for i := 0; i < death; i++ {
			seats = append(seats, opp)
		}
	}
	if len(seats) == 0 {
		return discard(ctx.Game, nil)
	}
	return ctx.Game.PlayersSacrificeThenForEffect(ctx.Source(), seats,
		sacrificeSpec("a creature", Creature()), "Capital Punishment — sacrifice a creature", discard)
})
