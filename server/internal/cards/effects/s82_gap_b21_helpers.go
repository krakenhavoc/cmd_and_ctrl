package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82GapB21DamageToBlocked is "It deals n damage to target creature it's
// blocking" for a creature that may have left the battlefield as the ability
// resolves (Tinder Wall sacrifices itself to pay): the damage is dealt by the
// source as it last existed (CR 608.2h), and only to a target still legal
// (CR 608.2b).
func S82GapB21DamageToBlocked(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ref, ok := ctx.SourceRef()
		if !ok {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard || t.ID == uuid.Nil {
				continue
			}
			return DealDamage{SourceObject: &ref, Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	}
}

// s82GapB21AnaxAsk is Anax and Cymede & Kynaios and Tiro's land clause for
// one player: "may put a land card from their hand onto the battlefield" (a
// put, not a play; that player puts it, CR 110.2a), then their answer
// decides the rest in s82GapB21AnaxTail.
func S82GapB21AnaxAsk(ctx *Context, player uuid.UUID) error {
	put := MayPutALandFromHand("Anax and Cymede & Kynaios and Tiro")
	put.Player = player
	put.ThenTail = s82GapB21AnaxTail
	put.ThenItem = true
	return ctx.Game.PutByEnteringPlayerForEffect(func() error { return put.Apply(ctx) })
}

// s82GapB21AnaxTail runs once one player has answered: an opponent of the
// controller who didn't put a land draws a card ("each opponent who didn't"),
// then the next player in turn order from the controller is asked. A keyed
// tail; an on-disk key, never renamed, never reused.
var s82GapB21AnaxTail game.TailRef

// (assigned in init: the tail and the ask refer to each other.)
func init() {
	s82GapB21AnaxTail = PutFromHandTail("s82-gap-b21/anax-kynaios-land-then-draw", func(ctx *Context, r PutFromHandResult) error {
		controller := ctx.Item.Controller
		if r.Player != controller && r.Entered == uuid.Nil {
			if err := (DrawCards{Player: r.Player, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		order := append([]uuid.UUID{controller}, OpponentsInTurnOrder(ctx.Game, controller)...)
		for i, p := range order {
			if p == r.Player && i+1 < len(order) {
				return S82GapB21AnaxAsk(ctx, order[i+1])
			}
		}
		return nil
	})
}
