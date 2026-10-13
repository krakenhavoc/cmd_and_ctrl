package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AnyPlayerPays is "<effect> unless any player pays <Cost>" and "Any
// player may pay <Cost>. If no one does, <effect>" (ADR 0150, #1903):
// the payment is offered to every player still in the game, in APNAP
// order from the active player, the controller included, until one
// pays (game.QueueAnyPlayerPaysForEffect). The offer blocks the step
// it was made in and takes no standing answer.
//
// Both branches are bound to the resolving stack item the way
// PayUnless's are, so OnNobody reads the item's targets and X when it
// finally runs (Wild Might's creature, Soul Strings' cards).
type AnyPlayerPays struct {
	// Cost is the payment ("{2}"), with any X already worked out.
	Cost string
	// Question is the prompt header each player sees.
	Question string
	// OnNobody is the effect that happens when nobody pays.
	OnNobody func(ctx *Context) error
	// OnPaid is "If a player does, …", handed the payer. Optional.
	OnPaid func(ctx *Context, payer uuid.UUID) error
	// HelpsController marks a payment the controller wants made (Icy
	// Prison's upkeep): game.AnyPlayerPays.HelpsController.
	HelpsController bool
}

func (a AnyPlayerPays) Apply(ctx *Context) error {
	item := ctx.Item
	var onPaid func(*game.Game, uuid.UUID) error
	if a.OnPaid != nil {
		paid := a.OnPaid
		onPaid = func(g *game.Game, payer uuid.UUID) error {
			return paid(NewContext(g, item), payer)
		}
	}
	return ctx.Game.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
		Source:     ctx.Source(),
		Controller: ctx.Controller(),
		Cost:       a.Cost,
		Question:   a.Question,
		OnPaid:     onPaid,
		OnNobody:   declineAgainstTheSameItem(item, a.OnNobody),

		HelpsController: a.HelpsController,
	})
}
