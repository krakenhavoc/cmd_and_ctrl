package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sail into the West — Instant {2}{G}{U}:
//
//	"Will of the council — Starting with you, each player votes for
//	 return or embark. If return gets more votes, each player returns
//	 up to two cards from their graveyard to their hand, then you
//	 exile Sail into the West. If embark gets more votes or the vote is
//	 tied, each player may discard their hand and draw seven cards."
//
// The vote is ADR 0146's (CR 701.38); a tie goes to embark, as printed.
//
// Return: every player, the caster first, chooses up to two cards from
// their own graveyard and returns them to hand. Sail into the West is
// still on the stack while it resolves, so it is not one of the cards
// anyone can return; "then you exile Sail into the West" is done before
// the choices open, which nobody can tell apart from after (the exile
// is not conditional on what was returned) and which keeps the spell
// out of every graveyard for good.
//
// Embark: every player is asked whether to discard their hand and draw
// seven; each answers for themselves, and a player who says no keeps
// their hand and draws nothing.
//
// A bot casting it votes embark; an opponent's bot votes return, since
// a refilled graveyard costs it nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a8956d0-bd37-4305-8443-4df043796f6e",
		Name:         "Sail into the West",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Sail into the West — vote for return or embark",
				Words:         []string{"return", "embark"},
				ForController: []int{1, 2},
				ForOpponents:  []int{2, 1},
				Then:          sailIntoTheWestVoted,
			}.Apply(ctx)
		},
	})
}

var sailIntoTheWestVoted = VoteResultThen("vote/sail-into-the-west", func(ctx *Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		if err := ctx.Game.ExileCardForEffect(r.Source); err != nil && err != game.ErrCardNotFound {
			return err
		}
		for _, p := range tablePlayers(ctx) {
			if err := (ReturnChosenFromGraveyard{
				Player:   p,
				Question: "Sail into the West — return up to two cards from your graveyard to your hand",
				Except:   r.Source,
				Min:      0,
				Max:      2,
				ToHand:   true,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	for _, p := range tablePlayers(ctx) {
		if err := embarkOffer(ctx, p); err != nil {
			return err
		}
	}
	return nil
})

// embarkOffer asks one player whether to discard their hand and draw
// seven cards.
func embarkOffer(ctx *Context, player uuid.UUID) error {
	return MayChoice{
		Player:   player,
		Question: "Sail into the West — discard your hand and draw seven cards?",
		Draws:    7,
		Then:     sailIntoTheWestAnswer,
	}.Apply(ctx)
}

// sailIntoTheWestEmbark is one player's "discard your hand and draw seven
// cards". Every player's question is open at once, so the steps find their
// player from the seat RunAt was given. The draw waits for a prompt a discard
// asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155).
var sailIntoTheWestEmbark = RegisterThenSteps("then-steps/sail-into-the-west-embark",
	Instruction(func(ctx *Context) error {
		_, err := discardWholeHand(ctx.Game, seatAtOccurrence(ctx))
		return err
	}),
	Instruction(func(ctx *Context) error {
		return DrawCards{Player: seatAtOccurrence(ctx), N: 7}.Apply(ctx)
	}),
)

// sailIntoTheWestAnswer is a player's "yes": they discard their hand and
// draw seven. Keyed (#4291).
var sailIntoTheWestAnswer = MayYes("confirm/sail-into-the-west-embark", func(ctx *Context, r game.ConfirmAnswer) error {
	return sailIntoTheWestEmbark.RunAt(ctx, seatIndexOf(ctx.Game, r.Chooser))
})
