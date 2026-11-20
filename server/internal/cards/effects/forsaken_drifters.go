package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forsaken Drifters — Creature — Zombie {3}{B}, 4/2 (Born of the Gods,
// #5559):
//
//	"When this creature dies, mill four cards."
//
// Aven Heartstabber's dies trigger with the mill alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e94f685-f392-40a8-8ad4-683351877fd9",
		Name:         "Forsaken Drifters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Forsaken Drifters — mill four cards", Do(MillCards{N: 4})),
		},
	})
}
