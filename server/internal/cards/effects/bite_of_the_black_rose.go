package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bite of the Black Rose — Sorcery {3}{B}:
//
//	"Will of the council — Starting with you, each player votes for
//	 sickness or psychosis. If sickness gets more votes, creatures your
//	 opponents control get -2/-2 until end of turn. If psychosis gets
//	 more votes or the vote is tied, each opponent discards two cards."
//
// The vote is ADR 0146's (CR 701.38); a tie goes to psychosis, as
// printed. Sickness shrinks the creatures your opponents control as it
// resolves (CR 611.2c: the set is fixed then). Psychosis asks each
// opponent to choose two cards from their own hand.
//
// A bot casting it votes psychosis; an opponent's bot votes sickness,
// since -2/-2 for a turn costs less than two cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4db04e23-e1da-4c41-9dba-5502dfff73fb",
		Name:         "Bite of the Black Rose",
		Completeness: CompletenessFull,
		// The sickness branch is a one-sided -2/-2 sweep (ADR 0126 §6);
		// the vote may go the other way, which the declaration cannot say.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Bite of the Black Rose — vote for sickness or psychosis",
				Words:         []string{"sickness", "psychosis"},
				ForController: []int{1, 2},
				ForOpponents:  []int{1, 0},
				Then:          biteOfTheBlackRoseVoted,
			}.Apply(ctx)
		},
	})
}

var biteOfTheBlackRoseVoted = VoteResultThen("vote/bite-of-the-black-rose", func(ctx *Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		return BoostUntilEOT{
			Match:     And(Creature(), OpponentControls()),
			Power:     -2,
			Toughness: -2,
			Label:     "Bite of the Black Rose — creatures your opponents control get -2/-2",
		}.Apply(ctx)
	}
	eachOpponentDiscards(ctx, 2)
	return nil
})
