package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// learn.go — the keyword action learn (CR 701.48, #3525), built on the
// outside-the-game seam (ADR 0191, ADR 0198).
//
// CR 701.48a: "Learn" means "You may discard a card. If you do, draw a
// card. If you didn't discard a card, you may reveal a Lesson card you
// own from outside the game and put it into your hand."
//
// The two halves are one instruction, so they are one question and then,
// for a player who declined the first, a second:
//
//  1. a discard prompt with a floor of zero ("you may discard a card"),
//     the same prompt Abandon Attachments asks. A player with an empty
//     hand is asked nothing and has "didn't discard a card", so learn
//     goes straight to the Lessons;
//  2. only for a player who discarded nothing, the Lesson pick: a
//     TakeFromOutside over the player's sideboard (or the card database
//     for a deck that seated none, ADR 0198), "you may", revealed. With
//     no Lesson to choose from it asks nothing.
//
// Learn is one event, opened once (game.LearnForEffect) as a
// RepEventKeywordAction before any of it happens, so "if you would
// learn, ... instead" can replace the whole instruction
// (retriever_phoenix.go, ADR 0191's amendment of 2026-10-23). A
// replaced learn discards, draws and reveals nothing.
//
// Card text writes "learn" as the last instruction of its ability in
// every printing, so the card files put it last in a ResolveSteps or a
// Do and nothing needs to wait on it; the walk still parks a later step
// behind either prompt, as it does for any other.
//
// The player discards before the Lesson is looked at. That order is not
// a UI choice: a player who discards a card has "discarded a card", and
// the Lesson half is then not taken at all.

// isLessonCard is the Lesson clause: a Lesson is a spell subtype
// (CR 205.3k), read off the card's type line like Arcane or Adventure.
func isLessonCard(c game.Card) bool { return c.HasSubtype("Lesson") }

// learnLessonQuestion is the Lesson prompt's header.
const learnLessonQuestion = "Learn — reveal a Lesson card you own from outside the game, or none"

// learnDiscardQuestion is the discard prompt's header.
const learnDiscardQuestion = "Learn — discard a card to draw a card, or none to look for a Lesson instead"

// Learn is "learn" (CR 701.48a). Player zero means the resolving
// effect's controller.
type Learn struct {
	Player uuid.UUID
}

func (l Learn) Apply(ctx *Context) error {
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	// Learn is an event a replacement effect can apply to (CR 614.1a,
	// game/learn_action.go): Retriever Phoenix, in a graveyard, may
	// replace it. Whatever the window leaves is learnBody.
	return ctx.Game.LearnForEffect(player, ctx.Source())
}

func init() { game.LearnBody = learnBody }

// learnBody is learn itself, once its CR 614 window has left it alone:
// CR 701.48a's discard, draw and Lesson.
func learnBody(g *game.Game, player, source uuid.UUID) error {
	return g.PlayerDiscardsThenForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   source,
		N:        1,
		UpTo:     true,
		Question: learnDiscardQuestion,
	}, func(g *game.Game, done game.PromptedDiscards) error {
		if done.Discarded(player) {
			// "If you do, draw a card."
			return g.DrawNForEffect(player, 1)
		}
		// "If you didn't discard a card, you may reveal a Lesson card
		// you own from outside the game and put it into your hand."
		takeFromOutside(g, player, source, TakeFromOutside{
			Match:    isLessonCard,
			Question: learnLessonQuestion,
			Reveal:   true,
		})
		return nil
	})
}

// OnTheTargetCard is one step of a spell that does something to its
// first legal target card and then learns: "Target creature gets +3/+1
// and gains trample until end of turn. Learn." is
//
//	ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { … }), Learn{})
//
// A target that has gone by resolution does nothing (CR 608.2b), and the
// steps after it still run: the spell was not countered, only that
// clause had nothing to act on. A spell whose only target is gone is
// countered whole by the engine before any step runs.
func OnTheTargetCard(f func(ctx *Context, id uuid.UUID) error) Instruction {
	return func(ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return f(ctx, id)
	}
}
