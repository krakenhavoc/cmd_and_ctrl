package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ridgeline Rager — Creature — Beast {2}{R}, 1/2 (Prophecy, #5617):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Firefly's firebreathing without the flying: the shared
// thisGetsUntilEndOfTurn body at {R}. No tap in the cost, so it can be
// activated any number of times, summoning sick or not, and the boost
// ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54e7a6d4-d405-4627-93ab-8d8bd945fe02",
		Name:         "Ridgeline Rager",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Ridgeline Rager — +1/+0"),
		}},
	})
}
