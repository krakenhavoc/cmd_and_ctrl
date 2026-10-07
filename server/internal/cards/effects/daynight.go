package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// daynight.go is the card half of day and night (CR 731, ADR 0132). The
// mechanic itself — the designation, the untap-step check and the
// daybound / nightbound turn-over — is engine (game/daynight.go). What a
// card needs from it is these:
//
//	BecomeDay{} / BecomeNight{} / ToggleDayNight{}   // "it becomes day"
//	BecomesDayAsEnters()                             // "If it's neither day nor night, it becomes day as ~ enters."
//	WheneverDayBecomesNightOrNightBecomesDay(…)      // the trigger half
//	ItsNight(g) / ItsDay(g)                          // "if it's night"
//
// A daybound / nightbound werewolf needs none of it: the keyword is
// stamped by the deck importer (or declared in PrintedKeywords) and the
// engine does the rest.

// BecomeDay is "it becomes day". Already day, it does nothing.
type BecomeDay struct{}

func (BecomeDay) Apply(ctx *Context) error {
	ctx.Game.BecomeDayForEffect()
	return nil
}

// BecomeNight is "it becomes night".
type BecomeNight struct{}

func (BecomeNight) Apply(ctx *Context) error {
	ctx.Game.BecomeNightForEffect()
	return nil
}

// ToggleDayNight is "If it's night, it becomes day. Otherwise, it
// becomes night." — The Celestus. Day or neither goes to night.
type ToggleDayNight struct{}

func (ToggleDayNight) Apply(ctx *Context) error {
	ctx.Game.ToggleDayNightForEffect()
	return nil
}

// BecomesDayAsEnters is the AsEnters slot for "If it's neither day nor
// night, it becomes day as ~ enters." A game that already has a
// designation is left alone. Off the stack, because the clause is an
// "as enters" one (CR 614.12) and nobody gets a window to answer it.
func BecomesDayAsEnters() func(card *game.Card, ctx *Context) error {
	return func(_ *game.Card, ctx *Context) error {
		ctx.Game.BecomeDayIfNeitherForEffect()
		return nil
	}
}

// DayBecameNightOrNightBecameDay is the condition "day becomes night or
// night becomes day" (CR 731.1a): the designation FLIPPED. The first
// designation a game gains is not one of those.
func DayBecameNightOrNightBecameDay(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return game.DayNightFlipped(ev)
}

// WheneverDayBecomesNightOrNightBecomesDay — "Whenever day becomes
// night or night becomes day, <effect>" (The Celestus, Brimstone
// Vandal, Firmament Sage). Every player's copy of the permanent
// triggers: the flip belongs to the game, not to a seat.
func WheneverDayBecomesNightOrNightBecomesDay(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDayNightChanged, DayBecameNightOrNightBecameDay, label, effect)
}

// ItsNight and ItsDay are the condition readers: "if it's night".
// Neither is true while the game has no designation. Caller holds g.mu,
// as every card callback does.
func ItsNight(g *game.Game) bool { return g.IsNight() }
func ItsDay(g *game.Game) bool   { return g.IsDay() }

// ItsNightCost is the cost-modifier condition "if it's night" —
// Moonrager's Slash's "This spell costs {2} less to cast if it's night."
// Read-only, under the cast path's write lock, like every CostPredicate.
func ItsNightCost() CostPredicate {
	return func(q game.CostQuery) bool { return q.Game != nil && q.Game.IsNight() }
}
