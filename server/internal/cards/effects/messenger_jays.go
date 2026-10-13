package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Messenger Jays — Creature — Bird {4}{U}, 2/1:
//
//	"Flying
//	 Council's dilemma — When this creature enters, starting with you,
//	 each player votes for feather or quill. Put a +1/+1 counter on
//	 this creature for each feather vote and draw a card for each quill
//	 vote. For each card drawn this way, discard a card."
//
// An enters trigger whose resolution is ADR 0146's council's dilemma
// (CR 701.38). Feather votes become counters on the Jays — the object
// that entered (CR 400.7), so a Jays that left while the trigger waited
// gets none. Quill votes draw that many cards, and only once every
// draw has settled does the controller discard the same number, so the
// drawn cards are legal discards (draw_then.go).
//
// The discard count is the cards that really arrived, not the quill
// votes: a draw a replacement effect swallowed (Notion Thief on the
// opponent's side of the table) costs no discard.
//
// A bot votes feather for its own Jays and quill for an opponent's,
// where the card filtering is the smaller gift.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dfeb1d36-d1d9-43dd-a79f-a98560676bb3",
		Name:            "Messenger Jays",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Messenger Jays — each player votes for feather or quill", voteForThisCreature(Vote{
				Question:      "Messenger Jays — vote for feather or quill",
				Words:         []string{"feather", "quill"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          messengerJaysVoted,
			})),
		},
	})
}

var messengerJaysVoted = VoteResultThen("vote/messenger-jays", func(ctx *Context, r game.VoteResult) error {
	if err := countersOnVotedCreature(ctx, r, r.Votes(0)); err != nil {
		return err
	}
	if n := r.Votes(1); n > 0 {
		return ctx.Game.DrawNThenForEffect(ctx.Controller(), n, game.DrawThen{
			Ref: messengerJaysDiscardRef, Player: ctx.Controller(), Source: ctx.Source(), N: n,
		})
	}
	return nil
})

// messengerJaysDiscardRef is "for each card drawn this way, discard a
// card": it runs once every draw has settled and asks for one discard per
// card that really arrived. A draw a replacement swallowed or redirected
// (Notion Thief, dredge) leaves no EventDrawCard for the player, so it
// costs no discard. The draws are counted back to the vote's last ballot,
// the last thing that happened before them.
var messengerJaysDiscardRef = game.RegisterDrawThen("messenger-jays-discard", func(g *game.Game, d game.DrawThen) error {
	drawn := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventVoteCast {
			break
		}
		if ev.Kind == game.EventDrawCard && ev.Actor == d.Player {
			drawn++
		}
	}
	drawn = min(drawn, d.N)
	if drawn > 0 {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player: d.Player, Source: d.Source, N: drawn,
			Question: "Messenger Jays — discard a card for each card drawn",
		})
	}
	return nil
})
