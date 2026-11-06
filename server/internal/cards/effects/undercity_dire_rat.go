package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercity Dire Rat — Creature — Rat {1}{B}, 2/2 (Final Fantasy, #4462):
//
//	"Rat Tail — When this creature dies, create a Treasure token."
//
// Impulsive Pilferer's dies trigger. "Rat Tail" is a flavor word with no
// rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f09811d0-c5da-4061-aaa7-835ccee3b05a",
		Name:         "Undercity Dire Rat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Undercity Dire Rat — Rat Tail — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}
