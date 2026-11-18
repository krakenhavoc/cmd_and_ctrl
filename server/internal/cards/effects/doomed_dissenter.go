package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomed Dissenter — Creature — Human {1}{B}, 1/1 (Reprint, #5908):
//
//	"When this creature dies, create a 2/2 black Zombie creature token."
//
// Doomed Traveler's shape: a mandatory dies trigger making one token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11cb509d-21af-48f1-b355-135ebd3e4bd1",
		Name:         "Doomed Dissenter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Doomed Dissenter — create a 2/2 Zombie", Do(CreateToken{
				Template: TokenCard("2/2 black Zombie"),
				N:        1,
			})),
		},
	})
}
