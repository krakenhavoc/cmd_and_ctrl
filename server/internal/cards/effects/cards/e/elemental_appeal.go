package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elemental Appeal — Sorcery {R}{R}{R}{R} (Zendikar, S82 OLD batch 491,
// #5583):
//
//	"Kicker {5}
//	 Create a 7/1 red Elemental creature token with trample and haste. Exile it
//	 at the beginning of the next end step. If this spell was kicked, that
//	 creature gets +7/+0 until end of turn."
//
// The token is created through a keyed tail that knows the token that landed:
// the +7/+0 goes on it if the spell was kicked (read from the spell's cast
// record carried by the tail), and the exile is a delayed trigger over it
// (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3bcfc82c-3523-40e3-b236-78ece8a8d69f",
		Name:          "Elemental Appeal",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{5}")},
		Purpose:       game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			kicked := 0
			if ctx.WasKicked() {
				kicked = 1
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("7/1 red Elemental with trample and haste")), Count: 1}},
			}, effects.WithN(ctx.ItemTail(elementalAppealToken), kicked))
		},
	})
}

var elementalAppealToken = effects.CardTail("zen/elemental-appeal-token", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	if o.Tail.N > 0 {
		if err := (effects.BoostUntilEOT{Target: o.IDs[0], Power: 7, Label: "Elemental Appeal — +7/+0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Elemental Appeal — exile the Elemental token",
		Cards:      append([]uuid.UUID(nil), o.IDs...),
		Body:       effects.ExileListedCardsBody,
	}.Apply(ctx)
})
