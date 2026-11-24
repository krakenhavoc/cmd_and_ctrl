package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ephara, God of the Polis — Legendary Enchantment Creature — God {2}{W}{U},
// 6/5 (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to white and blue is less than seven, Ephara
//	 isn't a creature.
//	 At the beginning of each upkeep, if you had another creature enter the
//	 battlefield under your control last turn, draw a card."
//
// The God clause is the shared Theros static (CR 700.5, a hybrid symbol of both
// colours is one symbol). "Each upkeep" is any player's, and the intervening
// "if" (CR 603.4) reads the PREVIOUS turn's record of creatures that entered
// under the source's controller (TurnTally.PrevEnteredCreatures), whoever's turn
// it was, excluding Ephara itself: checked as the upkeep begins and again as the
// trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5968ecb5-5728-429f-9324-0c108820ebe8",
		Name:            "Ephara, God of the Polis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "W", "U")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.EnteredCreatureLastTurnForEffect(source.Controller, source.InstanceID)
			}, "Ephara, God of the Polis — draw a card",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "you had another creature enter the battlefield under your control last turn", Test: func(ctx *effects.Context) bool {
						return ctx.Game.EnteredCreatureLastTurnForEffect(ctx.Controller(), ctx.Item.SourceCardID)
					}},
					Then: []effects.Applier{effects.DrawCards{N: 1}},
				})),
		},
	})
}
