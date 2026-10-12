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
// Expanded draws each open a replacement window and inherit the
// effects already applied to their parent. Two doublers therefore
// produce four cards, while each descendant independently consults
// first-draw exceptions, draw redirects, and dredge.

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
	// one you draw in each of your draw steps". The shared step tally
	// exempts only the first completed draw.
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
			if d.ExceptInOwnDrawStep && g.FirstDrawInOwnDrawStepForEffect(ev.DrawPlayer) {
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

// YouDrawTwiceInstead is "if you would draw a card, draw two cards
// instead" — Thought Reflection.
func YouDrawTwiceInstead(label string) game.ReplacementEffect {
	return DrawBecomes{Count: timesTwo, Scope: DrawsByController, Label: label}.Build()
}

// YouDrawTwiceInsteadExceptTheFirst is the same sentence with
// Alhammarret's Archive's "except the first one you draw in each of
// your draw steps".
func YouDrawTwiceInsteadExceptTheFirst(label string) game.ReplacementEffect {
	return DrawBecomes{
		Count:               timesTwo,
		Scope:               DrawsByController,
		ExceptInOwnDrawStep: true,
		Label:               label,
	}.Build()
}
