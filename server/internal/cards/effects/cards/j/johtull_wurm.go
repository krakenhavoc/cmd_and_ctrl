package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Johtull Wurm — Creature — Wurm {5}{G}, 6/6 (Ice Age, S82 OLD, #5636):
//
//	"Whenever this creature becomes blocked, it gets -2/-1 until end of turn for
//	 each creature blocking it beyond the first."
//
// Counted as the trigger resolves (CR 509.1h): one blocker is nothing.
//
// No simplification.
func init() {
	const label = "Johtull Wurm — gets -2/-1 until end of turn for each creature blocking it beyond the first"
	effects.Register(effects.Spec{
		OracleID:     "ab5a8242-4e42-43e6-a6ee-eeb0a8db3d11",
		Name:         "Johtull Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisBecomesBlocked, label, effects.OldB835JohtullWurm(label)),
		},
	})
}
