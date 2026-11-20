package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunter of Nightveil — Creature — Spirit {3}{U}{B}, 3/4:
//
//	"Creatures your opponents control get -1/-0."
//
// A layer 7c static (CR 613.4c) over every creature an opponent controls,
// Waker of Waves's sentence through the shared anthem body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "496c852b-ec62-4320-afbb-e19f534954c9",
		Name:         "Haunter of Nightveil",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(b16Anthem(b27CreaturesOpponentsControl, -1, 0), "Creatures your opponents control get -1/-0."),
		},
	})
}
