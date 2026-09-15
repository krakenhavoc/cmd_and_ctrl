package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b45fdeab-00cc-4422-af9f-66f30a880a7c",
		Name:         "Evolution Sage",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You don't choose what to proliferate — the game picks for you, adding every counter that helps you and every counter that hurts an opponent."},
		Triggered: []game.TriggeredAbility{
			Landfall("Evolution Sage — proliferate (landfall)", Do(Proliferate{})),
		},
	})
}
