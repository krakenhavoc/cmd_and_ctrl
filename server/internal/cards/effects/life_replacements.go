package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_replacements.go — the CR 614 replacement on the LIFE GAINED:
// "if you would gain life, you gain twice that much life instead"
// (Rhox Faithmender, Alhammarret's Archive's first half). Its own file
// rather than helpers.go, per the convention enters_tapped.go,
// mill_replacements.go and mana_replacements.go set.
//
// The event is game.RepEventLife, and nothing here is new engine work:
// since #482 every writer of a life total runs the CR 614 window and
// lands in one tail (game/life_tail.go), so a clause in this family
// doubles a catalog GainLife, a drain's gain half and the CR 702.15b
// lifelink credit alike. What this file adds is the second card on the
// family, which is what turned an inline clause into a shared one.
//
// DAMAGE is not this window. A Lightning Helix's three damage is
// neither reduced nor doubled by a life replacement (CR 120.3); only
// its three life of gain is.

// LifeGainBecomes is "if <Scope> would gain life, they gain
// <Amount(n)> life instead" — with the arithmetic as a function.
//
// Amount is applied to the delta the event currently carries, NOT to
// the printed one, which is the rules' own composition: CR 616.1
// applies one replacement and then re-gathers, so two DIFFERENT
// doublers in one window are ×4 with an ordering prompt, and two
// copies of the same card are ×4 without one (#792).
//
// It applies to a GAIN only — a positive delta. A life LOSS and a life
// PAYMENT go through untouched, which is what "gain life" says.
type LifeGainBecomes struct {
	Amount func(n int) int
	Scope  LifeGainScope
	Label  string
}

// LifeGainScope narrows whose life gain a replacement of this family
// watches.
type LifeGainScope string

const (
	// LifeGainByAnyone is the zero value: every player's gain. Nothing
	// prints it in this family yet — the honest zero rather than a
	// silent narrowing.
	LifeGainByAnyone LifeGainScope = ""

	// LifeGainByController is "if YOU would gain life" — Rhox
	// Faithmender, Alhammarret's Archive.
	LifeGainByController LifeGainScope = "controller"

	// LifeGainByOpponents is "if an OPPONENT would gain life". No
	// printed card in this family, but it is the third of three and
	// costs one arm.
	LifeGainByOpponents LifeGainScope = "opponents"
)

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (l LifeGainBecomes) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventLife || ev.LifeDelta <= 0 {
				return false
			}
			return l.scopeMatches(ev.LifePlayer, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if l.Amount == nil {
				return nil
			}
			ev.LifeDelta = l.Amount(ev.LifeDelta)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: l.Label,
	}
}

// scopeMatches answers "is this a gain this card cares about?".
func (l LifeGainBecomes) scopeMatches(gainer uuid.UUID, src *game.Card) bool {
	if src == nil || src.Controller == uuid.Nil {
		return false
	}
	switch l.Scope {
	case LifeGainByOpponents:
		return gainer != src.Controller
	case LifeGainByController:
		return gainer == src.Controller
	}
	return true
}

// YouGainTwiceThatMuchLife is "if you would gain life, you gain twice
// that much life instead" — Rhox Faithmender, Alhammarret's Archive.
func YouGainTwiceThatMuchLife(label string) game.ReplacementEffect {
	return LifeGainBecomes{Amount: timesTwo, Scope: LifeGainByController, Label: label}.Build()
}
