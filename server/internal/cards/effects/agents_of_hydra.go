package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agents of HYDRA — Creature — Human Spy Villain {1}{B}, 1/1
// (Marvel Super Heroes):
//
//	"When this creature dies, create a 2/1 black Villain creature token
//	 with menace."
//
// Doomed Traveler's dies trigger; the token is the table's 2/1 black
// Villain with menace row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31816924-1430-4581-86e8-325e2ca16217",
		Name:         "Agents of HYDRA",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Agents of HYDRA — create a 2/1 black Villain with menace", Do(CreateToken{
				Template: TokenCard("2/1 black Villain with menace"),
				N:        1,
			})),
		},
	})
}
