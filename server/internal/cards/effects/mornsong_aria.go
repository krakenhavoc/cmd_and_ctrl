package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mornsong Aria — Legendary Enchantment {1}{B}{B}:
//
//	"Players can't draw cards or gain life.
//	 At the beginning of each player's draw step, that player loses 3
//	 life, searches their library for a card, puts it into their hand,
//	 then shuffles."
//
// #3927. Maralen of the Mornsong with "can't gain life" beside the draw
// limit at zero: two statics, Spec.DrawLimits (CR 121.3) and
// Spec.CantGainLife (CR 119.7). The draw-step trigger is a search, not a
// draw (CR 121.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ca2f668-2a4d-4a1e-abcc-093ccbc29295",
		Name:         "Mornsong Aria",
		Completeness: CompletenessFull,
		DrawLimits:   PlayersCantDrawCards(),
		CantGainLife: PlayersCantGainLife(),
		Triggered:    []game.TriggeredAbility{mornsongDrawStep("Mornsong Aria")},
	})
}
