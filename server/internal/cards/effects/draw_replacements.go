package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_replacements.go — the CR 614 replacement on the DRAW AMOUNT:
// "if you would draw a card, draw two cards instead" (Thought
// Reflection, Alhammarret's Archive). Its own file rather than
// helpers.go, per the convention enters_tapped.go and
// mill_replacements.go set.
//
// The event is game.RepEventDraw, which has existed since S17 — what
// #1222 added to it is a COUNT. The base is always one, because
// CR 121.2 makes "draw three cards" three individual card draws and
// DrawNForEffect loops: a doubler therefore doubles EACH of them (a
// Thought Reflection on "draw three" draws six), and a redirect
// (Notion Thief) can still take one card of three rather than all or
// nothing.
//
// What a doubled draw does NOT do is re-open the window. The two cards
// are performed one at a time, so every per-card payoff — Nekusar,
// Sheoldred, Consecrated Sphinx, Fate Unraveler — still fires per
// card, but they are ONE event and CR 614.5's once-per-event tracking
// covers both. That is what makes two Thought Reflections draw four
// rather than three: the second multiplies what the first left,
// through the ordinary CR 616.1 apply-loop.
//
// The redirect family (Notion Thief's "you draw that card instead")
// lives on the same event and needs nothing from this file: it
// rewrites DrawPlayer, this rewrites DrawCount, and the two compose
// through the apply-loop in the order the drawing player picks.

// DrawBecomes is "if <Scope> would draw a card, they draw <Count(n)>
// cards instead" — the whole family, with the arithmetic as a
// function.
//
// Count is applied to the count the event currently carries, NOT to
// the printed one, which is the rules' own composition: CR 616.1
// applies one replacement and then re-gathers, so a doubler and a
// second doubler in one window are ×4. Two copies of the SAME card do
// not ask — they are two objects contributing one declared effect, and
// #792's identical-window skip applies — so two Alhammarret's Archives
// draw four with no prompt.
type DrawBecomes struct {
	Count func(n int) int

	// Scope narrows whose draws this watches.
	Scope DrawScope

	// ExceptInOwnDrawStep is Alhammarret's Archive's "except the first
	// one you draw in each of your draw steps". See
	// drawnInOwnDrawStep: the engine keeps no per-draw-step tally, so
	// this reads as "except ANY draw during that player's own draw
	// step", which is the same declared simplification Notion Thief
	// ships with.
	ExceptInOwnDrawStep bool

	Label string
}

// DrawScope narrows whose draws a replacement of this family watches.
type DrawScope string

const (
	// DrawsByAnyone is the zero value: every player's draw. Nothing
	// prints it in this family yet — it is the honest zero rather than
	// a silent narrowing.
	DrawsByAnyone DrawScope = ""

	// DrawsByController is "if YOU would draw" — Thought Reflection,
	// Alhammarret's Archive.
	DrawsByController DrawScope = "controller"

	// DrawsByOpponents is "if an OPPONENT would draw". No printed card
	// in this family, but it is the third of three and costs one arm.
	DrawsByOpponents DrawScope = "opponents"
)

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (d DrawBecomes) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDrawCard},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 {
				// Already replaced down to nothing: "would draw a card"
				// is the printed condition and there is no draw left.
				return false
			}
			if g.PlayerByIDForEffect(ev.DrawPlayer) == nil {
				return false
			}
			if d.ExceptInOwnDrawStep && drawnInOwnDrawStep(g, ev.DrawPlayer) {
				return false
			}
			return d.scopeMatches(ev.DrawPlayer, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if d.Count == nil {
				return nil
			}
			ev.DrawCount = d.Count(ev.DrawCount)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: d.Label,
	}
}

// scopeMatches answers "is this a draw this card cares about?".
func (d DrawBecomes) scopeMatches(drawer uuid.UUID, src *game.Card) bool {
	if src == nil || src.Controller == uuid.Nil {
		return false
	}
	switch d.Scope {
	case DrawsByOpponents:
		return drawer != src.Controller
	case DrawsByController:
		return drawer == src.Controller
	}
	return true
}

// drawnInOwnDrawStep reads "except the first one you draw in each of
// your draw steps" as "during that player's own draw step at all".
//
// THE DECLARED SIMPLIFICATION, and it is Notion Thief's, word for
// word: the engine keeps no per-draw-step draw tally, so a SECOND draw
// in somebody's draw step — an instant cast there, a draw-step trigger
// — is left alone rather than caught. For a card that takes something
// away from the drawer (Notion Thief) that is weaker than printed; for
// a card that gives them something (Alhammarret's Archive) it is
// weaker too, because the extra card the Archive would have given on
// that second draw does not arrive. Both directions cost the
// replacement's controller, never their opponents, which is the
// posture every declared simplification in the catalog takes.
//
// Closing it is a per-draw-step tally on the player — the second half
// of the "Draw-replacement count" seam row, and the one Teferi's
// Ageless Insight also waits on.
func drawnInOwnDrawStep(g *game.Game, drawer uuid.UUID) bool {
	if g.Turn.Step != game.StepDraw {
		return false
	}
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	seat := g.Seats[g.Turn.ActiveSeat]
	return seat != nil && seat.ID == drawer
}

// YouDrawTwiceInstead is "if you would draw a card, draw two cards
// instead" — Thought Reflection.
func YouDrawTwiceInstead(label string) game.ReplacementEffect {
	return DrawBecomes{Count: timesTwo, Scope: DrawsByController, Label: label}.Build()
}

// YouDrawTwiceInsteadExceptTheFirst is the same sentence with
// Alhammarret's Archive's "except the first one you draw in each of
// your draw steps" — read as "except any draw in your own draw step".
// See drawnInOwnDrawStep.
func YouDrawTwiceInsteadExceptTheFirst(label string) game.ReplacementEffect {
	return DrawBecomes{
		Count:               timesTwo,
		Scope:               DrawsByController,
		ExceptInOwnDrawStep: true,
		Label:               label,
	}.Build()
}
