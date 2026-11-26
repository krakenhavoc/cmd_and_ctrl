package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB288IntoHandAndGraveyard settles "put that pile into your hand and the
// other into your graveyard" (Epiphany at the Drownyard): the chosen pile
// moves to the caster's hand as a keyed bounce, and the other pile is its
// tail, binned through rest.
func OldB288IntoHandAndGraveyard(ctx *Context, rest game.TailRef, toHand, toGraveyard []uuid.UUID) error {
	tail := WithCarry(ctx.Tail(rest), toGraveyard...)
	if len(toHand) == 0 {
		return tail.Run(ctx.Game, nil)
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(toHand, tail)
}

// OldB288BinCarried is the tail that puts each carried card still in a zone
// into its owner's graveyard.
func OldB288BinCarried(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
