package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleam of Battle — Enchantment {4}{R}{W} (Dragon's Maze, S82 OLD b405,
// #5564):
//
//	"Whenever a creature you control attacks, put a +1/+1 counter on it."
//
// Fervent Charge's trigger: one per attacking creature you control, read off
// the declaration's EventAttack (CR 508.1m), and "it" is the creature the
// event names. A creature that has left by resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f23d9e-0c05-4028-9c64-60b9a20020f9",
		Name:         "Gleam of Battle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Gleam of Battle — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
