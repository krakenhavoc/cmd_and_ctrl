package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cyclone Summoner — Creature — Giant Wizard {5}{U}{U}, 7/7:
//
//	"When this creature enters, if you cast it from your hand, return all
//	 permanents to their owners' hands except for Giants, Wizards, and
//	 lands."
//
// Breaching Leviathan's intervening "if" (CR 603.4) over the cast record:
// a Summoner put onto the battlefield, reanimated or cast from anywhere but
// the hand does nothing, checked as it enters and again as the trigger
// resolves. The permanents returned are those on the battlefield as it
// resolves that are none of a land, a Giant or a Wizard, read through the
// layered view; tokens are returned and cease to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2088e791-9ce5-4e97-a4ce-f51adaa3f42b",
		Name:         "Cyclone Summoner",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepBounce}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Cyclone Summoner — return all permanents to their owners' hands except for Giants, Wizards, and lands",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B30CastFromHand(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.BounceAllMatching{
						Match: effects.And(effects.Nonland(), effects.Not(effects.HasSubtype(ctx.TextSubtype("Giant"))), effects.Not(effects.HasSubtype(ctx.TextSubtype("Wizard")))),
					}.Apply(ctx)
				}),
		},
	})
}
