package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thieving Otter — Creature — Otter {2}{U}, 2/2 (Bloomburrow, #4467):
//
//	"Whenever this creature deals damage to an opponent, draw a card."
//
// Looter il-Kor's trigger (WheneverThisDealsDamageToAnOpponent) with a
// plain draw: any damage this creature deals to a player other than its
// controller, in combat or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5dee663f-9e2e-4503-8086-fc51664fdf29",
		Name:         "Thieving Otter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Thieving Otter — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
