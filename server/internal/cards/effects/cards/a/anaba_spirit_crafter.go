package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anaba Spirit Crafter — Creature — Minotaur Shaman {2}{R}{R}, 1/3
// (Homelands, #5635):
//
//	"Minotaur creatures get +1/+0."
//
// A layer 7c anthem (CR 613.4c) over every Minotaur on the battlefield,
// whoever controls it. It does not say "other", and the Spirit Crafter
// is a Minotaur, so it pumps itself too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9bee79c-144f-41e4-a53d-17df8381f691",
		Name:         "Anaba Spirit Crafter",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Minotaur"}}, 1, 0),
		},
	})
}
