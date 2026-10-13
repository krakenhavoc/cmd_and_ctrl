package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// finish_voting.go — "Whenever players finish voting, …" (CR 701.38,
// ADR 0146 amendment 2026-10-13, #2927).
//
// The engine emits game.EventVotesFinished once per vote, carrying every
// ballot in the order cast. The trigger's stack item keeps that event, so
// the ability reads who voted for what as it RESOLVES — a vote cast
// earlier in the turn, or a second vote started in response, cannot
// change it. These helpers are the printed words:
//
//	Triggered: []game.TriggeredAbility{WheneverPlayersFinishVoting("Grudge Keeper — …",
//		func(g *game.Game, item *game.StackItem) error {
//			ctx := NewContext(g, item)
//			for _, opp := range OpponentsWhoVotedForAChoiceYouDidntVoteFor(ctx) { … }
//		})},
//
// Two readings are the whole difficulty, and both are about CR 701.38d's
// extra votes and about a player who cast no ballot:
//
//   - "A choice you voted for" is any of the choices the controller voted
//     for. A controller who voted twice names two choices; one who cast
//     nothing names none, so nobody voted for a choice they voted for.
//   - "A choice you didn't vote for" is any choice outside that set, so a
//     controller who cast nothing leaves every choice an opponent voted
//     for qualifying. An opponent who voted for both kinds is counted
//     under both clauses — they did vote for each.
//
// "Each opponent" is each opponent still in the game, whoever the vote's
// controller was: the trigger fires for a vote an opponent called too.

// WheneverPlayersFinishVoting is "Whenever players finish voting, …".
// It fires for every vote at the table, whoever called it (the ability
// is not "your" votes), once per vote — a vote with extra ballots
// (CR 701.38d) is still one vote. A vote that never began because there
// was nothing to vote for does not trigger it.
func WheneverPlayersFinishVoting(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventVotesFinished, func(game.Event, *game.Card, game.Characteristic, *game.Game) bool {
		return true
	}, label, effect)
}

// FinishedVoteBallots is every ballot of the vote the resolving
// "players finish voting" trigger is about, in the order cast. Empty
// for an item that is not such a trigger.
func FinishedVoteBallots(ctx *Context) []game.VoteBallot {
	return ctx.Trigger().Event.Ballots
}

// OpponentsWhoVotedForAChoiceYouVotedFor lists, in seat order, the
// opponents of the trigger's controller who voted for at least one
// choice the controller also voted for.
func OpponentsWhoVotedForAChoiceYouVotedFor(ctx *Context) []uuid.UUID {
	ballots, me := FinishedVoteBallots(ctx), ctx.Controller()
	var out []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if game.VotedForAChoiceVotedForBy(ballots, opp, me) {
			out = append(out, opp)
		}
	}
	return out
}

// OpponentsWhoVotedForAChoiceYouDidntVoteFor lists, in seat order, the
// opponents of the trigger's controller who voted for at least one
// choice the controller did not vote for.
func OpponentsWhoVotedForAChoiceYouDidntVoteFor(ctx *Context) []uuid.UUID {
	ballots, me := FinishedVoteBallots(ctx), ctx.Controller()
	var out []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if game.VotedForAChoiceNotVotedForBy(ballots, opp, me) {
			out = append(out, opp)
		}
	}
	return out
}
