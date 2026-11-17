package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrowing Razormaw — Creature — Beast {2}{G}, 4/2 (The Brothers' War):
//
//	"When this creature dies, mill four cards."
//
// One dies trigger (WhenThisDies) whose effect is the controller milling
// four, from the library as it is when the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8551e1df-47eb-419e-8dd2-e96b9a450912",
		Name:         "Burrowing Razormaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Burrowing Razormaw — mill four cards", Do(MillCards{N: 4})),
		},
	})
}
