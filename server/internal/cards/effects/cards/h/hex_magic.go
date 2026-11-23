package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hex Magic — {2}{R} Sorcery — Arcane (Marvel Super Heroes, #4455):
//
//	"Exile all the cards from your hand, then draw that many cards. Until
//	 the end of your next turn, you may play cards exiled this way."
//
// The hand leaves as one event (Herigast's), and "that many" is the number
// of cards that really reached exile, so a commander sent to the command
// zone instead is not counted. Each card that landed gets a "play" permission
// (a land is playable, not only castable) that ends with the end of your next
// turn (ADR 0063), granted before the draw. With an empty hand nothing is
// exiled and nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2283729c-11db-471c-9f63-46092c97bd00",
		Name:         "Hex Magic",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.ExileCardsKeyedForEffect(effects.AllHandCardIDs(ctx.Game, ctx.Controller()), ctx.ItemTail(hexMagicDraw))
		},
	})
}

// hexMagicDraw is the hand's exile finished: the permission over each card
// that landed, then one draw for each. An on-disk key.
var hexMagicDraw = effects.CardTail("hex-magic/permit-and-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	g, controller := ctx.Game, ctx.Controller()
	cards := make([]game.Card, 0, len(o.IDs))
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:   controller,
		Zone:     game.ZoneExile,
		Duration: g.UntilEndOfYourNextTurnDuration(controller),
	}, cards)
	return effects.DrawCards{Player: controller, N: len(o.IDs)}.Apply(ctx)
})
