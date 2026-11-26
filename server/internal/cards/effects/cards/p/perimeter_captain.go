package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perimeter Captain — Creature — Human Soldier {W} 0/4 (Worldwake, S82 OLD
// batch 489, #5582):
//
//	"Defender
//	 Whenever a creature you control with defender blocks, you may gain 2
//	 life."
//
// One trigger per blocking creature (CR 509.3a), the blocker's defender read
// through the layers (itself included). The life gain is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05608055-d97a-4c8f-833d-47b3dd1ea255",
		Name:            "Perimeter Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Purpose:         game.Purpose{LifeGain: 2},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBlock, effects.S82OldB489ADefenderYouControlBlocked,
				"Perimeter Captain — you may gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: 2}.Apply(effects.NewContext(g, item))
				}),
				"Perimeter Captain — gain 2 life?"),
		},
	})
}
