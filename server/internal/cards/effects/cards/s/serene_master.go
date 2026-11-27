package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serene Master — Creature — Human Monk {1}{W}, 0/2 (Commander 2013, S82 older
// sets):
//
//	"Whenever this creature blocks, exchange its power and the power of target
//	 creature it's blocking until end of combat."
//
// Both powers are read as the trigger resolves (layers applied, counters and
// pumps included), then each creature has its base power set to the other's
// until the end of the combat (CR 511.3), as Evra, Halcyon Witness sets a base
// power. Later pumps still apply on top. The target is a creature this Master
// is blocking (CR 509.1), judged again as it resolves; a Master or attacker
// that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ce0d583-81ca-4dca-bde0-52f86b683afd",
		Name:         "Serene Master",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.SelfBlocksOnce(ev, source)
				}, "Serene Master — exchange its power and the power of target creature it's blocking until end of combat", effects.OldB384ExchangePowerWithTarget),
				effects.BlockedBySource(effects.TargetCreature("target creature it's blocking"))),
		},
	})
}
