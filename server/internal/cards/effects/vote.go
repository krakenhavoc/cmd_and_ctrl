package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vote.go — voting, the card side (CR 701.38, ADR 0146). The engine
// half is game/council_vote.go: one option pick per ballot, in turn
// order from the named player, with the extra votes of CR 701.38d, each
// vote logged as it is cast.
//
// A card writes "starting with you, each player votes for <A> or <B>"
// as a Vote and everything printed after the vote as a registered
// continuation (VoteResultThen), which reads the tally:
//
//	var pleaForPowerResult = VoteResultThen("vote/plea-for-power", func(ctx *Context, r game.VoteResult) error {
//		if r.MoreVotes(0, 1) { … time … }
//		… knowledge, or the vote is tied …
//	})
//
// Will of the council asks "which option got more votes"
// (VoteResult.MoreVotes, ties go to the option the card names for
// them), council's dilemma asks "how many votes did each get"
// (VoteResult.Votes, VotersFor), and an object vote asks "which got the
// most or tied for most" (VoteResult.MostVotes).

// Vote is "starting with you, each player votes for …".
type Vote struct {
	// Question is every ballot's header, "<card> — vote for <A> or
	// <B>".
	Question string

	// Words are the options when they are words printed on the card,
	// in printed order (CR 701.38b). Set Words or Options, not both.
	Words []string

	// Options are the options when they are objects or players, built
	// as the vote starts (VoteForPermanents). Labels are what the
	// voters read.
	Options []game.ChoiceOption

	// ForController and ForOpponents are the bot's hints: how much the
	// vote's controller, and an opponent of theirs, wants each option
	// to win, one number per option, higher is better. Nil leaves a
	// bot to judge object options by the objects and to take the
	// first word.
	ForController, ForOpponents []int

	// Then is the rest of the card, run with the tally.
	Then game.VoteThen

	// Carry is plain data Then is handed (r.Carry): the targets the
	// card printed before the vote, say.
	Carry []uuid.UUID
}

// Apply starts the vote, with the resolving effect's controller voting
// first ("starting with you"). Everything printed after the vote is
// Then: Apply returns with the first ballot queued.
func (v Vote) Apply(ctx *Context) error {
	opts := v.Options
	if len(v.Words) > 0 {
		opts = make([]game.ChoiceOption, len(v.Words))
		for i, w := range v.Words {
			opts[i] = game.ChoiceOption{Label: w}
		}
	}
	if len(opts) == 0 {
		// Nothing to vote for (no permanent qualified): nobody votes,
		// and the rest of the card reads an empty tally.
		return ctx.Game.RunVoteThenForEffect(v.Then, game.VoteResult{
			Controller: ctx.Controller(), Source: ctx.Source(), Carry: v.Carry,
		})
	}
	return ctx.Game.StartVoteForEffect(game.VotePrompt{
		Starting:      ctx.Controller(),
		Controller:    ctx.Controller(),
		Source:        ctx.Source(),
		Question:      v.Question,
		Options:       opts,
		ForController: v.ForController,
		ForOpponents:  v.ForOpponents,
		Then:          v.Then,
		Carry:         v.Carry,
	})
}

// VoteResultThen registers a vote's continuation, as a card file's
// package-level var. The body gets a Context rebuilt from values — the
// vote's controller as its controller, the card that called the vote as
// its source — and the tally. The key is an on-disk identity
// ("vote/<card>"): never renamed, never reused.
func VoteResultThen(key string, body func(ctx *Context, r game.VoteResult) error) game.VoteThen {
	return game.RegisterVoteThen(key, func(g *game.Game, r game.VoteResult) error {
		return body(NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   r.Controller,
			Owner:        r.Controller,
			SourceCardID: r.Source,
		}), r)
	})
}

// VoteForPermanents is an object vote's option list: one option per
// permanent, labelled with its name, in battlefield order. Caller holds
// g.mu.
func VoteForPermanents(g *game.Game, ids []uuid.UUID) []game.ChoiceOption {
	out := make([]game.ChoiceOption, 0, len(ids))
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		out = append(out, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{id}})
	}
	return out
}

// VotedPermanent is the permanent option i names, or uuid.Nil.
func VotedPermanent(r game.VoteResult, i int) uuid.UUID {
	if i < 0 || i >= len(r.Options) || len(r.Options[i].Cards) == 0 {
		return uuid.Nil
	}
	return r.Options[i].Cards[0]
}
