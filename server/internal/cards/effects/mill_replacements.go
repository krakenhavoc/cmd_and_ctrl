package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_replacements.go — the CR 614 replacement on the MILL AMOUNT:
// "if an opponent would mill one or more cards, they mill twice that
// many cards instead" (Bruvac the Grandiloquent) and "...they mill that
// many cards plus four instead" (The Water Crystal). Its own file
// rather than helpers.go, per the convention enters_tapped.go set.
//
// The event is game.RepEventMill (#569), opened once per mill
// INSTRUCTION before any card leaves the library — the way
// RepEventCreateTokens is opened once per creation and
// RepEventKeywordAction once per action. It carries the player milling
// and ONE field a replacement rewrites: the count.
//
// The per-card window is a different, older one and needs nothing from
// this file: every milled card routes through the shared exit
// primitive, so "if a card would be put into a graveyard from anywhere,
// exile it instead" and CR 903.9 already see each of them.
// graveyard_replacements.go is where that family lives.

// MillScope narrows WHOSE mills a replacement of this family watches.
// Both printed cards say "an opponent"; the zero value is every
// player's mill, so a symmetrical printing is the field left out.
type MillScope string

const (
	// MillsByAnyone is the zero value: every player's mill, the
	// source's controller included. Nothing prints it yet — it is the
	// honest zero rather than a silent narrowing.
	MillsByAnyone MillScope = ""

	// MillsByOpponents is "if an OPPONENT would mill" — Bruvac the
	// Grandiloquent, The Water Crystal.
	MillsByOpponents MillScope = "opponents"

	// MillsByController is "if YOU would mill". No printed card, but it
	// is the third of three and costs one arm.
	MillsByController MillScope = "controller"
)

// MillBecomes describes "if <Scope> would mill one or more cards, they
// mill <Count(n)> cards instead" — the whole family, with the
// arithmetic as a function.
//
// Count is applied to the count the event currently carries, NOT to the
// printed one, and that is the rules' own composition: CR 616.1 applies
// one replacement and then re-gathers, so a doubler and a "plus four"
// in the same window compose in the order the MILLING PLAYER picks (×2
// then +4 is 10 from a base of 3; +4 then ×2 is 14). Two copies of the
// same card do not ask — they are two objects contributing one declared
// effect, and #792's identical-window skip applies.
//
// The count is the one the INSTRUCTION named, not what the library can
// supply: CR 701.13b's "mill as many as possible" clamp happens after
// this window settles, so Bruvac doubles a mill of twenty against a
// twelve-card library.
type MillBecomes struct {
	Count func(n int) int
	Scope MillScope
	Label string
}

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (m MillBecomes) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventMill},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMill || ev.MillCount <= 0 {
				// Already replaced down to nothing, or never a count:
				// "one or more cards" is the printed condition and
				// there is no mill left to replace.
				return false
			}
			return m.scopeMatches(ev.MillPlayer, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if m.Count == nil {
				return nil
			}
			ev.MillCount = m.Count(ev.MillCount)
			return nil
		},
		// The EFFECT's controller, which is not the affected player and
		// on both printed cards is deliberately somebody else: CR 616.1
		// gives the ordering choice to the player milling, which
		// affectedPlayerForEvent reads off the event (#982).
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: m.Label,
	}
}

// scopeMatches answers "is this a mill this card cares about?".
//
// Anyone-else rather than a seat-by-seat opponent check, the posture
// OpponentsSpell takes: the game has no teams, so every other player at
// a Commander table is an opponent, and a source with no controller
// (a fixture, a token mid-construction) matches nobody rather than
// everybody.
func (m MillBecomes) scopeMatches(miller uuid.UUID, src *game.Card) bool {
	switch m.Scope {
	case MillsByOpponents:
		return src != nil && src.Controller != uuid.Nil && miller != src.Controller
	case MillsByController:
		return src != nil && src.Controller != uuid.Nil && miller == src.Controller
	}
	return true
}

// OpponentsMillTwice is "if an opponent would mill one or more cards,
// they mill twice that many cards instead" — Bruvac the Grandiloquent.
func OpponentsMillTwice(label string) game.ReplacementEffect {
	return MillBecomes{Count: timesTwo, Scope: MillsByOpponents, Label: label}.Build()
}

// OpponentsMillPlus is "if an opponent would mill one or more cards,
// they mill that many cards plus n instead" — The Water Crystal's four.
// "That many" is the count the event carries, which is why this adds to
// it rather than setting it.
func OpponentsMillPlus(n int, label string) game.ReplacementEffect {
	return MillBecomes{
		Count: func(c int) int { return c + n },
		Scope: MillsByOpponents,
		Label: label,
	}.Build()
}
