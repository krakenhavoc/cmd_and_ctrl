package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate.go — the Proliferate primitive (CR 701.34):
//
//	"Choose any number of permanents and/or players with counters on
//	 them, then give each another counter of each kind already
//	 there."
//
// The rule is one choice and one application. game.ProliferateForEffect
// is the application — and, since #976, the CR 614 window on the
// keyword ACTION, so "if you would proliferate, proliferate twice
// instead" (Tekuthal, Inquiry Dominus) is a replacement the primitive
// needs to know nothing about.
//
// THE CHOICE IS THE PLAYER'S (#2525). Proliferate{} with no lists asks
// game.ProliferateChoosingForEffect, which queues a
// PendingChoiceProliferate once the window has settled: a multi-select
// over every permanent and player that has a counter, with the
// engine's beneficial pick suggested (game.ProliferateSuggestionForEffect
// — everything of yours a counter helps, everything of theirs one
// hurts). That is what a player takes nearly every time, so it is the
// default, but it is no longer the only answer: a player can now
// decline a counter — "put another -1/-1 counter on my own persist
// creature to stop it coming back", or leave a Stun counter off a
// permanent — which the old auto-pick could not.
//
// Because the choice is a prompt, a proliferate PAUSES the effect that
// asked for it. Anything the card says after "proliferate" goes in
// Proliferate.Then, which runs once every proliferate of the settled
// count has been answered (or at once when there is nothing to choose),
// never on the next line of the card's OnResolve — see Steady Progress,
// whose "draw a card" would otherwise be drawn while the player is
// still deciding.

// Proliferate gives each chosen permanent and player one more counter
// of each kind already on it.
//
// With Cards and Players both nil the proliferating player is asked.
// A card that knows exactly what it wants (or a test) sets them
// explicitly and no prompt is queued; an explicitly empty, non-nil
// slice means "choose nothing", which is a legal proliferate.
type Proliferate struct {
	// Controller is the player proliferating. Zero value means the
	// effect's controller, which is what every printed proliferate
	// wants.
	Controller uuid.UUID

	// Cards / Players are an explicit choice. Nil for both means
	// "ask the player".
	Cards   []uuid.UUID
	Players []uuid.UUID

	// Then is the rest of the sentence after "proliferate". It runs
	// with the game lock held, once the proliferate is finished.
	// Capture scalars only (the controller's ID), never a *Context or
	// a *Game: an undone game resolves it against the restored one.
	Then func(g *game.Game) error
}

func (p Proliferate) Apply(ctx *Context) error {
	controller := p.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	if p.Cards == nil && p.Players == nil {
		return ctx.Game.ProliferateChoosingForEffect(controller, ctx.Source(), p.Then)
	}
	if err := ctx.Game.ProliferateForEffect(controller, ctx.Source(), p.Cards, p.Players); err != nil {
		return err
	}
	if p.Then != nil {
		return p.Then(ctx.Game)
	}
	return nil
}

// BeneficialProliferateChoice is the pick a proliferate suggests on
// `controller`'s behalf (game.ProliferateSuggestionForEffect). Kept
// under this name for the tests and for any card that wants to hand a
// beneficial pick to Proliferate.Cards / Players explicitly.
//
// Only things that already have at least one counter can be chosen
// (CR 701.34a), so an empty board or a board with no counters on it
// returns two empty lists and the proliferate is a legal no-op.
func BeneficialProliferateChoice(g *game.Game, controller uuid.UUID) (cards []uuid.UUID, players []uuid.UUID) {
	return g.ProliferateSuggestionForEffect(controller)
}
