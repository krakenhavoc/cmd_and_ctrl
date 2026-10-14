package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// vote.go — a ballot (CR 701.38, ADR 0146): an option_pick that carries
// `council_vote`.
//
// Two kinds of vote, two rules.
//
//   - Words ("time or knowledge"). The card says on the ballot how much
//     the vote's controller, and an opponent of theirs, wants each word
//     to win (`for_controller`, `for_opponents`). The bot votes for the
//     word its side wants most. A card that gives no hint gets the first
//     word, the enumerator's own preference.
//   - Objects ("a nonland permanent you don't control", Council's
//     Judgment). Every shipped object vote is removal, so the bot votes
//     for the most valuable permanent it doesn't control, and against
//     its own, leaning a little towards one that already has votes so a
//     table of bots does not split four ways.
//
// An extra vote it may decline ("Don't vote again") is worth half of
// its favourite: it always casts it, for the option it likes best, and
// would rather decline than vote for its least favourite.
//
// A secret vote (ADR 0180) is scored by the same rules. The bot reads
// its seat's filtered view, where a secret vote's tally is all zeros
// and only its own ballots are listed, so the lean towards an option
// with votes does nothing and no other seat's ballot can steer it. A
// vote for "up to one" offers a vote for nothing, which it casts only
// when nothing else is on offer. With `self_best` (Círdan) it votes
// for itself.

// voteValue scores answering ballot `ch` with offered option `index`.
// The second return is false for an option pick that is not a ballot.
func (st *state) voteValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || ch.CouncilVote == nil || index == nil {
		return 0, "", false
	}
	v := ch.CouncilVote
	i := *index
	if i < 0 || i >= len(v.Offered) || i >= len(ch.PickOptions) {
		return 0, "", false
	}
	opt := v.Offered[i]
	if opt == -2 {
		// "Up to one" (ADR 0180): a vote for nothing is the last resort.
		return 0.1, "vote for nothing", true
	}
	if opt < 0 {
		return 0.5, "decline the extra vote", true
	}
	if v.SelfBest && ch.PickOptions[i].Player != "" {
		// ADR 0180: the votes a player receives are good for them
		// (Círdan the Shipwright), so every player votes for themselves.
		if ch.PickOptions[i].Player == st.me {
			return 2, "vote for ourselves", true
		}
		return 0.25, "vote: " + v.Options[opt], true
	}
	hints := v.ForOpponents
	if v.Controller == st.me {
		hints = v.ForController
	}
	if len(hints) == len(v.Options) && len(hints) > 0 {
		lo, hi := hints[0], hints[0]
		for _, h := range hints {
			lo, hi = min(lo, h), max(hi, h)
		}
		if hi == lo {
			return 1, "vote: " + v.Options[opt], true
		}
		return float64(hints[opt]-lo) / float64(hi-lo), "vote for what our side wants: " + v.Options[opt], true
	}
	cards := ch.PickOptions[i].Cards
	if len(cards) == 0 {
		// Words with no hint: the first one.
		if opt == 0 {
			return 1, "vote: " + v.Options[opt], true
		}
		return 0.25, "vote: " + v.Options[opt], true
	}
	c := &cards[0]
	if live := st.bf[c.InstanceID]; live != nil {
		c = live
	}
	tally := 0
	if opt < len(v.Tally) {
		tally = v.Tally[opt]
	}
	if v.Zone != "" {
		// Cards in a graveyard (Custodi Squire): they return to the
		// vote's controller, so its bot votes for the best card and an
		// opponent's for the worst, both leaning towards a card that
		// already has votes.
		if v.Controller == st.me {
			return 1 + st.permanentValue(c) + 0.5*float64(tally), "vote for our best card in the " + v.Zone + ": " + c.Name, true
		}
		return 1 - st.permanentValue(c) + 0.5*float64(tally), "vote for their worst card in the " + v.Zone + ": " + c.Name, true
	}
	if c.Controller == st.me {
		return -st.permanentValue(c), "vote against our own " + c.Name, true
	}
	return 1 + st.permanentValue(c) + 0.5*float64(tally), "vote for their best permanent: " + c.Name, true
}
