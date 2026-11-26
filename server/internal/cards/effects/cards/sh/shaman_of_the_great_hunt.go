package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaman of the Great Hunt — {3}{R} Creature — Orc Shaman 4/2 (Fate Reforged):
//
//	"Haste
//	 Whenever a creature you control deals combat damage to a player, put a
//	 +1/+1 counter on it.
//	 Ferocious — {2}{G/U}{G/U}: Draw a card for each creature you control with
//	 power 4 or greater."
//
// The trigger is Bident of Thassa's shape: one per creature per damaged
// player, and the counter goes on the creature that dealt the damage, if it is
// still on the battlefield. Ferocious is an ability word with no condition of
// its own; the draw counts the creatures as the ability resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f87eb3c-c261-4d92-a260-4b20daddb601",
		Name:            "Shaman of the Great Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Shaman of the Great Hunt — put a +1/+1 counter on the creature that dealt the damage",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.Source) {
						return nil
					}
					return effects.AddCounter{Target: item.Trigger.Event.Source, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Ferocious — {2}{G/U}{G/U}: Draw a card for each creature you control with power 4 or greater.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G/U}{G/U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.Msc13CreaturesWithPowerAtLeast(g, item.Controller, 4)
				return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
