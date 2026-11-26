package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foster — Enchantment {2}{G}{G} (Commander 2013, S82 OLD b380, #5560):
//
//	"Whenever a creature you control dies, you may pay {1}. If you do, reveal
//	 cards from the top of your library until you reveal a creature card. Put
//	 that card into your hand and the rest into your graveyard."
//
// Bitter Chill's "you may pay {1}" on a creature-you-control-died trigger
// (WheneverACreatureYouControlDies); the paid branch reveals until a creature
// card (Oath of Druids' RevealUntil), takes it to hand, and mills the rest. A
// library with no creature card puts everything revealed into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb431500-152c-4524-b76b-de62922ff57f",
		Name:         "Foster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Foster — you may pay {1}; if you do, reveal until a creature card, put it into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Foster — pay {1} to reveal cards until a creature card?",
						Then:     fosterPaid,
					}.Apply(effects.NewContext(g, item))
				}), game.Purpose{DeathPayoff: true}),
		},
	})
}

// fosterPaid is the paid branch. An on-disk key: never renamed, never reused.
var fosterPaid = effects.PayUnlessBranches("pay-unless/foster-reveal-until-creature",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		player := ctx.Controller()
		run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() }, "Foster")
		var rest []uuid.UUID
		for _, id := range run {
			if id != hit {
				rest = append(rest, id)
			}
		}
		if hit != uuid.Nil {
			ctx.Game.TakeFromLibraryToHandForEffect(player, []uuid.UUID{hit})
		}
		return effects.RestIntoGraveyard(ctx.Game, rest)
	}, nil)
