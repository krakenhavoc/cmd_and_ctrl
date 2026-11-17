package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capashen Templar — Creature — Human Knight {2}{W}, 2/2 (Urza's
// Destiny, #5621):
//
//	"{W}: This creature gets +0/+1 until end of turn."
//
// Honor Guard's +0/+1 at {W} on a 2/2, with the shared
// thisGetsUntilEndOfTurn body: no {T} in the cost, so it can be
// activated any number of times, summoning sick or not, and every boost
// ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab9ad5e2-2bf3-450d-9d5c-c7280d368e10",
		Name:         "Capashen Templar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGetsUntilEndOfTurn(0, 1, "Capashen Templar — +0/+1 until end of turn"),
		}},
	})
}
