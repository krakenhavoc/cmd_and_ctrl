package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howling Golem — Artifact Creature — Golem {3}, 2/3 (Commander Legends):
//
//	"Whenever this creature attacks or blocks, each player draws a card."
//
// Carrion Rats's attacks-or-blocks trigger (it triggers once when it blocks,
// however many creatures it blocks) over Mikokoro's "each player draws a
// card", round the table from the controller (CR 121.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75269a31-3819-4410-8bb7-10de16810a5e",
		Name:         "Howling Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Howling Golem — each player draws a card", effects.EachPlayerDrawsACard),
		},
	})
}
