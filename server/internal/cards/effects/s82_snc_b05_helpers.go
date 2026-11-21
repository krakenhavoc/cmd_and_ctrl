package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b05_helpers.go — Audacious Swap (SNC/NCC, S82 batch 5): "The
// owner of target nonenchantment permanent shuffles it into their
// library, then exiles the top card of their library. If it's a land
// card, they put it onto the battlefield. Otherwise, they may cast it
// without paying its mana cost." The OWNER shuffles, exiles, puts and
// casts, so the free-cast offer goes to them, not to the spell's
// controller. Append-only.

func AudaciousSwapResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
		WithCarry(ctx.Tail(audaciousSwapShuffleThenExile), c.Owner))
}

var audaciousSwapShuffleThenExile = game.RegisterTail("then/audacious-swap-shuffle-exile", func(g *game.Game, o game.TailOutcome) error {
	owner := TailCarry0(o)
	if owner == uuid.Nil {
		return nil
	}
	if err := g.ShuffleLibraryForEffect(owner); err != nil {
		return err
	}
	ctx := TailContext(g, o.Tail)
	tail := ctx.ItemTail(audaciousSwapExiled)
	tail.Carry = []uuid.UUID{owner}
	return g.MillToZoneKeyedForEffect(owner, 1, game.ZoneExile, game.MillUntil{}, tail)
})

var audaciousSwapExiled = CardTail("then/audacious-swap-exiled", func(ctx *Context, o game.TailOutcome) error {
	owner := TailCarry0(o)
	if owner == uuid.Nil {
		return nil
	}
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c.IsLand() {
			if _, err := ctx.Game.ReturnFromExileToBattlefieldForEffect(id, owner, false); err != nil {
				return err
			}
			continue
		}
		perm := ExileUntilNonlandFreeCastPermission(ctx, "Audacious Swap")
		perm.Player = owner
		perm.Zone = game.ZoneExile
		ctx.Game.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
		OfferResolutionCastOver(ctx.Game, owner, perm.Source, perm.SourceName,
			game.MayCastKeywordFreeCast, DeclineLeaveInExile, id)
	}
	return nil
})

// drawXStep and loseXStep are Cut of the Profits' "You draw X cards and you
// lose X life", one step each.
func DrawXStep() Instruction {
	return func(ctx *Context) error {
		if x := ctx.X(); x > 0 {
			return DrawCards{Player: ctx.Controller(), N: x}.Apply(ctx)
		}
		return nil
	}
}

func LoseXStep() Instruction {
	return func(ctx *Context) error {
		if x := ctx.X(); x > 0 {
			return GainLife{Player: ctx.Controller(), Amount: -x}.Apply(ctx)
		}
		return nil
	}
}
