package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sam Kirk, Xenoanthropologist — {3}{U} Legendary Creature — Human
// Scientist 3/3 (Star Trek Commander, #4974):
//
//	"Whenever an opponent casts a creature spell, investigate. (Create a
//	 Clue token. It's an artifact with "{2}, Sacrifice this token: Draw
//	 a card.")"
//
// Investigate is creating the printed Clue token (CR 701.16a). The
// trigger reads the spell as cast, through Sam's text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eec55307-8525-4dae-a030-d210f3895138",
		Name:         "Sam Kirk, Xenoanthropologist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCast(Creature()),
				"Sam Kirk, Xenoanthropologist — investigate",
				Do(Investigate{N: 1})),
		},
	})
}
