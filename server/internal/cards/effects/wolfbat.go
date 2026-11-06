package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolfbat — Creature — Wolf Bat {2}{B}, 2/2:
//
//	"Flying
//	 Whenever you draw your second card each turn, you may pay {B}. If
//	 you do, return this card from your graveyard to the battlefield
//	 with a finality counter on it. (If a creature with a finality
//	 counter on it would die, exile it instead.)"
//
// The trigger watches from the graveyard only (InGraveyard), on
// Knights of Dol Amroth's "second card" condition: the turn's draw
// tally read after the draw, so it fires on the second draw and on no
// other, and a Wolfbat that reaches the graveyard after the second
// draw waits for the next turn (the rulings). Only real draws count.
//
// The {B} is a "you may pay" on resolution (MayPay); paid, the card
// returns under its owner's control with a finality counter as it
// enters (Excava's entry option), and the engine's CR 122.1h
// replacement exiles it instead of letting it die. A Wolfbat that has
// left the graveyard, or is a new object there, stays where it is.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "adc47a25-daab-49c7-93a5-5bbf2bc20cf9",
		Name:            "Wolfbat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Wolfbat — you may pay {B} to return it from your graveyard with a finality counter",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{B}",
						Question: "Wolfbat — pay {B} to return it from your graveyard to the battlefield with a finality counter?",
						Then:     wolfbatPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// wolfbatPaid is "If you do, return this card from your graveyard to
// the battlefield with a finality counter on it."
var wolfbatPaid = PayUnlessBranches("pay-unless/wolfbat-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		self := ctx.Source()
		if z := ctx.Game.FindCardZoneForEffect(self); z == nil || z.Kind != game.ZoneGraveyard || ctx.isNewSourceObjectAsThis(self) {
			return nil
		}
		return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{self}, game.ZoneEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterFinality: 1},
		}, game.TailRef{}, 0)
	}, nil)
