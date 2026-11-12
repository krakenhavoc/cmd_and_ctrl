package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion Howler — Creature — Zombie Wolf {3}{B}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Pay 1 life: This creature gets +2/-1 until end of turn."
//
// Life is the whole cost (CR 119.4: refused without the life to pay), and
// there is no limit on activations, so it can trade toughness for power
// until the -1s kill it (CR 704.5f). A trade is a pump (ADR 0142's
// rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e139176f-483f-4f05-af3d-0176d0048b62",
		Name:         "Carrion Howler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Pay 1 life: This creature gets +2/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    PayLife(1),
			Effect:  thisGetsUntilEndOfTurn(2, -1, "Carrion Howler — +2/-1 until end of turn"),
		}},
	})
}
