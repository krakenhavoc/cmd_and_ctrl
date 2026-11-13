package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain of Acid — Sorcery {3}{G} (Onslaught):
//
//	"Destroy target noncreature permanent. Then that permanent's
//	 controller may copy this spell and may choose a new target for that
//	 copy."
//
// The Chain cycle's self-copy (chain_of_vapor.go, CR 707.10) with no
// price on it. The permanent's controller is read BEFORE the destruction,
// since a card in a graveyard has no controller to ask, and the offer
// rides the destruction's continuation (Boseiju, Who Endures's shape), so
// it comes after the move has settled. The offer is not gated on the
// permanent dying: it is a clause about a PLAYER, so an indestructible
// permanent survives and its controller may still copy. The copy is
// theirs (CR 707.10) and they choose its new target (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b1a27bd-bb98-44f8-8357-666fabfeabf0",
		Name:         "Chain of Acid",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target noncreature permanent", Noncreature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID},
				withCarry(ctx.ItemTail(chainOfAcidDestroyed), c.Controller))
		},
	})
}

// chainOfAcidDestroyed offers the copy to the destroyed permanent's
// controller (Carry[0]) once the destruction has settled. Keyed (#4253);
// an on-disk key: never renamed, never reused.
var chainOfAcidDestroyed = CardTail("then/chain-of-acid-offer-copy", func(ctx *Context, o game.TailOutcome) error {
	controller := tailCarry0(o)
	if controller == uuid.Nil {
		return nil
	}
	return chainOfferCopy(ctx.Game, ctx.Item, controller, "Chain of Acid")
})
