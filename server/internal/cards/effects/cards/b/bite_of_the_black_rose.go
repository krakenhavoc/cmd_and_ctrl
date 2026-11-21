package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4db04e23-e1da-4c41-9dba-5502dfff73fb",
		Name:         "Bite of the Black Rose",
		Completeness: effects.CompletenessFull,
		// The sickness branch is a one-sided -2/-2 sweep (ADR 0126 §6);
		// the vote may go the other way, which the declaration cannot say.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Vote{
				Question:      "Bite of the Black Rose — vote for sickness or psychosis",
				Words:         []string{"sickness", "psychosis"},
				ForController: []int{1, 2},
				ForOpponents:  []int{1, 0},
				Then:          biteOfTheBlackRoseVoted,
			}.Apply(ctx)
		},
	})
}

var biteOfTheBlackRoseVoted = effects.VoteResultThen("vote/bite-of-the-black-rose", func(ctx *effects.Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		return effects.BoostUntilEOT{
			Match:     effects.And(effects.Creature(), effects.OpponentControls()),
			Power:     -2,
			Toughness: -2,
			Label:     "Bite of the Black Rose — creatures your opponents control get -2/-2",
		}.Apply(ctx)
	}
	effects.EachOpponentDiscards(ctx, 2)
	return nil
})
