package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chance-Met Elves — Creature — Elf Warrior {2}{G}, 3/2 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"Whenever you scry, put a +1/+1 counter on this creature. This ability
//	 triggers only once each turn."
//
// An EventScry trigger the controller caused (Galadriel of Lothlórien's,
// Prudent Fateseer's), limited by the engine's per-object, per-turn
// "triggers only once each turn" record (TriggersOnlyOnceEachTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8958a6fe-76ef-45e3-8522-6bc13486b447",
		Name:         "Chance-Met Elves",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.On(game.EventScry, effects.ByYou,
				"Chance-Met Elves — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
