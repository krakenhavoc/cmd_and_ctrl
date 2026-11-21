package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ecstatic Beauty — Sorcery {2}{R} (Doctor Who):
//
//	"Exile the top three cards of your library. You may play those cards
//	 until end of turn. Put four time counters on each of those cards that
//	 has suspend.
//	 Suspend 4—{R}"
//
// The exile is keyed (a prompt it opens is a restore point). Its tail grants
// one play permission over the cards that reached exile, until end of turn
// (lands included, so it is a play and not only a cast), then puts four time
// counters on each of them that has suspend (printed suspend: a card exiled
// this way cannot have been granted it), which suspends it. Suspend is the
// engine's special action (CR 702.62).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "724d933c-6e5f-458c-9b39-83b4662693ea",
		Name:           "Ecstatic Beauty",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Suspend(4, "{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 3, ctx.Tail(ecstaticBeautyExiled))
		},
	})
}

// ecstaticBeautyExiled is the exile's tail. An on-disk key.
var ecstaticBeautyExiled = CardTail("ecstatic-beauty/play-and-time-counters", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	var cards []game.Card
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   g.UntilEndOfTurnDuration(),
		Source:     ctx.Source(),
		SourceName: "Ecstatic Beauty",
		Label:      "Ecstatic Beauty — you may play the exiled cards this turn",
	}, cards)
	for _, c := range cards {
		if !game.CardHasSuspend(c) {
			continue
		}
		if err := g.AddCounterByForEffect(ctx.Controller(), c.InstanceID, game.CounterTime, 4); err != nil {
			return err
		}
	}
	return nil
})
