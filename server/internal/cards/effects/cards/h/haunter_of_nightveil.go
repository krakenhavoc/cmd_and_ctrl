package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunter of Nightveil — Creature — Spirit {3}{U}{B}, 3/4:
//
//	"Creatures your opponents control get -1/-0."
//
// A layer 7c static (CR 613.4c) over every creature an opponent controls,
// Waker of Waves's sentence through the shared anthem body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "496c852b-ec62-4320-afbb-e19f534954c9",
		Name:         "Haunter of Nightveil",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(effects.B27CreaturesOpponentsControl, -1, 0), "Creatures your opponents control get -1/-0."),
		},
	})
}
