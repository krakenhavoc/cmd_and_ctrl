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
	return MillToZone{
		Player: ctx.Controller(),
		To:     game.ZoneExile,
		Until:  UntilCard(func(c game.Card) bool { return !c.IsLand() }),
		Then: func(ctx *Context, exiled []uuid.UUID) error {
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
			}
			return nil
		},
	}.Apply(ctx)
}
