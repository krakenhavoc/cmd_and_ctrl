package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const ridersOfTheMarkLabel = "Riders of the Mark — return it to its owner's hand; if you do, create 1/1 Human Soldier tokens equal to its toughness"

// Riders of the Mark — Creature — Human Knight {6}{R}, 7/4 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Affinity for Humans
//	 Trample, haste
//	 At the beginning of your end step, if this creature attacked this turn,
//	 return it to its owner's hand. If you do, create a number of 1/1 white
//	 Human Soldier creature tokens equal to its toughness."
//
// Amanda Grayson's affinity cost modifier over Humans, printed trample and
// haste, and Indoctrination Attendant's bounce with an "if you do" tail.
// The "if" is intervening (CR 603.4): this object attacked this turn, as
// the step begins and as the ability resolves. The tokens equal its
// toughness as it last stood on the battlefield (CR 608.2h), read before
// it leaves and carried on the tail.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7031047-5109-482f-b2e9-14800b9279bf",
		Name:            "Riders of the Mark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Humans", effects.HasSubtype("Human")),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this creature attacked this turn",
				Test: func(ctx *effects.Context) bool {
					return effects.OnBattlefield(ctx.Game, ctx.Source()) && ctx.Game.TimesAttackedThisTurn(ctx.Source()) > 0
				},
			}, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && g.TimesAttackedThisTurn(source.InstanceID) > 0
				},
				ridersOfTheMarkLabel,
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.BounceToHand{Target: c.InstanceID, ThenTail: effects.WithN(ctx.ItemTail(effects.GapB126RidersTail), c.CurrentToughness())}.Apply(ctx)
				})),
		},
	})
}
