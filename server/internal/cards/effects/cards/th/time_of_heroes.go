package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time of Heroes — Enchantment {1}{W} (Rise of the Eldrazi, S82 OLD batch
// 482, #5581):
//
//	"Each creature you control with a level counter on it gets +2/+2."
//
// A layer 7c anthem over the creatures its controller controls that have at
// least one level counter right now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcfb96d7-9b3b-4f60-8714-eebe4d9c002b",
		Name:         "Time of Heroes",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Counters[game.CounterLevel] > 0
			}, 2, 2),
		},
	})
}
