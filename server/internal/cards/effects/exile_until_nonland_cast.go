package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_until_nonland_cast.go — "exile cards from the top of your library
// until you exile a nonland card. You may cast that card by <price>
// rather than paying its mana cost." Shared by Amped Raptor (energy, ADR
// 0129 §5) and The Infamous Cruelclaw (a discard, ADR 0135 §2).
// Append-only (the clone gate's rule).

// exileUntilNonlandGrantingCast exiles from the top of the controller's
// library until a nonland card is exiled, then grants `perm` to that
// card if it is still in exile. The caller fills in the price, timing,
// lapse and labels; Player and Zone are set here.
func exileUntilNonlandGrantingCast(ctx *Context, perm game.CastPermission) error {
	// #4253 slice 10a: the permission rides the run's keyed tail as data.
	t := ctx.ItemTail(exileUntilNonlandGranted)
	t.Permission = &perm
	return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 0, game.ZoneExile, UntilNonland, t)
}

// exileUntilNonlandCastFree is "Exile cards from the top of your library
// until you exile a nonland card. You may cast that card without paying its
// mana cost." (Nico Minoru, Runaway; Wanda's Vision): the exile loop with a
// flash-speed, cast-only permission over the nonland card, priced {0}, made
// during the resolution (CR 608.2g) and lapsing with the card still in
// exile. An {X} in its cost is 0 (CR 107.3b).
func exileUntilNonlandCastFree(ctx *Context, sourceName string) error {
	return exileUntilNonlandGrantingCast(ctx, game.CastPermission{
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		Duration:    ctx.Game.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  sourceName,
		Label:       "Cast without paying its mana cost (" + sourceName + ")",
	})
}

// exileUntilNonlandGranted is the run's "then", told the cards exiled.
var exileUntilNonlandGranted = CardTail("exile-until-nonland/grant-cast", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.Permission == nil {
		return nil
	}
	return exileUntilNonlandGrant(ctx, *o.Tail.Permission, o.IDs)
})

// exileUntilNonlandGrant grants perm over the nonland card exiled.
func exileUntilNonlandGrant(ctx *Context, perm game.CastPermission, exiled []uuid.UUID) error {

	var granted []uuid.UUID
	for _, id := range exiled {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if !ok || c.IsLand() || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		grant := perm
		grant.Player = ctx.Controller()
		grant.Zone = game.ZoneExile
		ctx.Game.GrantCastPermissionToCardsForEffect(grant, []game.Card{c})
		granted = append(granted, id)
	}
	// #4250: cast during the resolution (CR 608.2g), at the
	// permission's price.
	offerResolutionCastOver(ctx.Game, ctx.Controller(), perm.Source, perm.SourceName,
		game.MayCastKeywordPaidCast, declineLeaveInExile, granted...)
	return nil
}
