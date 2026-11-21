package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolution Sage — Creature — Elf Druid {2}{G}, 3/2:
//
//	"Landfall — Whenever a land you control enters, proliferate."
//
// Lotus Cobra's trigger with a different payload: every land drop,
// every fetchland crack, every Cultivate is a free proliferate. The
// filter is the shared landfall helper, so "a land YOU control"
// excludes an opponent's land drop and includes a land that entered
// under your control from anywhere — played, fetched or reanimated.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b45fdeab-00cc-4422-af9f-66f30a880a7c",
		Name:         "Evolution Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Evolution Sage — proliferate (landfall)", effects.Do(effects.Proliferate{})),
		},
	})
}
