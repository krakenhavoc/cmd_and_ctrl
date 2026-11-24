package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oloro, Ageless Ascetic — Legendary Creature — Giant Soldier {3}{W}{U}{B},
// 4/5 (Secret Lair Drop):
//
//	"At the beginning of your upkeep, you gain 2 life.
//	 Whenever you gain life, you may pay {1}. If you do, draw a card and each
//	 opponent loses 1 life.
//	 At the beginning of your upkeep, if Oloro is in the command zone, you gain
//	 2 life."
//
// Three separate abilities. The first two work on the battlefield; the third
// is the command-zone twin (the eminence
// watch from the command zone only), and its intervening "if" (CR 603.4) is checked as
// the trigger resolves, so an Oloro that has been cast in response gains
// nothing. Each life gain, from either upkeep trigger, offers the payment: the
// payment is made as the trigger resolves, and the draw and the life loss
// happen only if {1} was really paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "620ff5f2-7d3f-467f-943d-3b62c2135023",
		Name:         "Oloro, Ageless Ascetic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Oloro, Ageless Ascetic — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
			effects.WheneverYouGainLife("Oloro, Ageless Ascetic — you may pay {1}: draw a card and each opponent loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Oloro, Ageless Ascetic — pay {1} to draw a card and make each opponent lose 1 life?",
						Then:     oloroPaid,
					}.Apply(effects.NewContext(g, item))
				}),
			oloroCommandZoneUpkeep(),
		},
	})
}

var oloroPaid = effects.PayUnlessBranches("pay-unless/oloro-draw-and-drain", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.EachOpponentLosesLife{N: 1}.Apply(ctx)
}, nil)

// oloroCommandZoneUpkeep is "At the beginning of your upkeep, if Oloro is in
// the command zone, you gain 2 life": watched from the command zone only.
func oloroCommandZoneUpkeep() game.TriggeredAbility {
	t := effects.AtYourUpkeep("Oloro, Ageless Ascetic — you gain 2 life (from the command zone)",
		effects.Do(effects.If{
			Cond: effects.CondBoard{Name: "Oloro is in the command zone", Test: func(ctx *effects.Context) bool {
				z := ctx.Game.FindCardZoneForEffect(ctx.Item.SourceCardID)
				return z != nil && z.Kind == game.ZoneCommand
			}},
			Then: []effects.Applier{effects.GainLife{Amount: 2}},
		}))
	t.Zones = []game.ZoneKind{game.ZoneCommand}
	return t
}
