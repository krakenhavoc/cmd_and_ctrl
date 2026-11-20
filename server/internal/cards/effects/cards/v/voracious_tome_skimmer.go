package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Tome-Skimmer — Creature — Faerie Rogue {U/B}{U/B}{U/B}, 2/3
// (Lorwyn Eclipsed, #4458):
//
//	"Flying
//	 Whenever you cast a spell during an opponent's turn, you may pay 1
//	 life. If you do, draw a card."
//
// The cast trigger is Katara, Waterbending Master's condition: the caster
// is the Tome-Skimmer's controller and is not the active player. "You may
// pay 1 life. If you do" is a payment (Seymour Flux's MayPay with
// LifePayment(1)), offered as the trigger resolves: a player who can't
// pay life (CR 119.4) is not asked, and the card is drawn only once the
// life is really paid.
//
// No simplification.
func init() {
	cast := effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
	}, "Voracious Tome-Skimmer — you may pay 1 life to draw a card",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Pay:      effects.LifePayment(1),
				Question: "Voracious Tome-Skimmer — pay 1 life to draw a card?",
				Then:     voraciousTomeSkimmerPaid,
			}.Apply(effects.NewContext(g, item))
		})
	cast.Purpose = game.Purpose{Draws: 1}
	effects.Register(effects.Spec{
		OracleID:        "25b0876e-c995-486e-8536-6e68b76b9c86",
		Name:            "Voracious Tome-Skimmer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// voraciousTomeSkimmerPaid is "draw a card", run once the life is paid.
// An on-disk key: never renamed, never reused.
var voraciousTomeSkimmerPaid = effects.PayUnlessBranches("pay-unless/voracious-tome-skimmer-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
