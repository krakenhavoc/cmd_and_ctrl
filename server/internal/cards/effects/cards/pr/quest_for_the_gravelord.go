package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quest for the Gravelord — Enchantment {B} (Battlebond, #5529):
//
//	Whenever a creature dies, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.
//
// The trigger is Blade of the Bloodchief's "a creature died" read, made optional; the activation is Khalni Heart Expedition's cost with a Zombie Giant token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5f92560-dc82-4bd2-ae6b-4c3942d458b7",
		Name:         "Quest for the Gravelord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died
			}, "Quest for the Gravelord — put a quest counter", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(effects.NewContext(g, item))
			}), "Quest for the Gravelord — put a quest counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 black Zombie Giant"), N: 1}),
		}},
	})
}
