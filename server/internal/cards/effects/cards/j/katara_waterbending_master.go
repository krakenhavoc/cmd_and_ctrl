package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katara, Waterbending Master — Legendary Creature — Human Warrior
// Ally {1}{U}, 1/3:
//
//	"Whenever you cast a spell during an opponent's turn, you get an
//	 experience counter.
//	 Whenever Katara attacks, you may draw a card for each experience
//	 counter you have. If you do, discard a card."
//
// The cast trigger is Brineborn Cutthroat's condition — the caster is
// Katara's controller and is not the active player — paying out in
// the player counter Aang, Airbending Master also uses
// (youGetAnExperienceCounter).
//
// The attack trigger's "you may" is asked at RESOLUTION (MayChoice,
// #796), where the card prints it, and the count is read there too,
// so a counter gained in response is included. The discard is linked
// to the draw (CR 607.2, Mask of Memory's shape): it happens only on
// the branch where the cards were drawn, and the player picks it from
// the hand the draws landed in. With no experience counters there is
// nothing to draw, so nothing is asked and nothing is discarded —
// drawing zero cards is not having drawn.
//
// No simplification.
func init() {
	const (
		castLabel   = "Katara, Waterbending Master — you get an experience counter"
		attackLabel = "Katara, Waterbending Master — draw a card per experience counter, then discard"
	)
	effects.Register(effects.Spec{
		OracleID:     "2e3af413-6706-4023-871a-e5e49da5beab",
		Name:         "Katara, Waterbending Master",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
			}, castLabel, effects.YouGetAnExperienceCounter),
			effects.WheneverThisAttacks(attackLabel, kataraMayDrawPerExperience),
		},
	})
}

// kataraMayDrawPerExperience is "you may draw a card for each
// experience counter you have. If you do, discard a card."
//
// Caller holds g.mu.
func kataraMayDrawPerExperience(g *game.Game, item *game.StackItem) error {
	n := effects.ExperienceCounters(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return effects.MayChoice{
		Question: "Katara, Waterbending Master — draw a card for each experience counter, then discard a card?",
		Draws:    n,
		YesLabel: "Draw",
		NoLabel:  "Decline",
		Then:     kataraAnswer,
	}.Apply(effects.NewContext(g, item))
}

// kataraAnswer is the "yes": draw a card for each experience counter, then
// discard a card. Keyed (#4291).
var kataraAnswer = effects.MayYes("confirm/katara-draw-discard", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.B16DrawThenDiscard(ctx.Game, ctx.Item, effects.ExperienceCounters(ctx.Game, ctx.Controller()))
})
