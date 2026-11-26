package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b529_helpers.go — shared pieces for the S82 OLD batch 529 cards
// (Shadowmoor life gain and loss, #5591).

// OldB529WhenAPlayerCasts is "whenever a player casts a <color> spell", any
// player's cast, the source's controller included (Kor Firewalker's shape).
func OldB529WhenAPlayerCasts(color, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCast, TmcB8APlayerCast(OfColor(color)), label, effect)
}

// OldB529MayPayOne is the "you may pay {1}. If you do, ..." the Initiates
// print: the controller is asked as the ability resolves (CR 603.5, 118.12),
// and `then` runs only once the {1} is really paid. `carry` is plain data the
// paid branch is handed as PayUnlessAnswer.Carry (the Smolder Initiate's target).
func OldB529MayPayOne(question string, then game.PayUnlessElse, carry ...uuid.UUID) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{1}",
			Question: question,
			Then:     then,
			Carry:    carry,
		}.Apply(NewContext(g, item))
	}
}

// OldB529PlayerHasNoCardsInHand is "that player has no cards in hand" for the
// player whose upkeep it is (the trigger event's actor), CR 603.4: checked as
// the ability triggers and again as it resolves. A hand's size is public.
type OldB529PlayerHasNoCardsInHand struct{}

func (OldB529PlayerHasNoCardsInHand) Holds(ctx *Context) bool {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return false
	}
	return OldB529HandIsEmpty(ctx.Game, ctx.Item.Trigger.Event.Actor)
}

func (OldB529PlayerHasNoCardsInHand) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// OldB529HandIsEmpty reports a seated player with no cards in hand.
func OldB529HandIsEmpty(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && (p.Hand == nil || p.Hand.Size() == 0)
}

// OldB529TappedArtifactCreatureAndLandCount is Toil to Renown's count: the
// tapped artifacts, creatures and lands `player` controls, a permanent that is
// two of them counted once.
func OldB529TappedArtifactCreatureAndLandCount(g *game.Game, player uuid.UUID) int {
	return CountControlled(g, player, func(c game.Card) bool {
		return c.Tapped && (c.IsArtifact() || c.IsCreature() || c.IsLand())
	})
}
