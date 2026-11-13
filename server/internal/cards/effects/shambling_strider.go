package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Strider — Creature — Yeti {4}{G}{G}, 5/5 (Ice Age, #5636):
//
//	"{R}{G}: This creature gets +1/-1 until end of turn."
//
// Flowstone Shambler's +1/-1 at {R}{G}, with the shared
// thisGetsUntilEndOfTurn body (selfPumpRow is for non-negative sizes
// only). Five activations leave it 10/0, and it dies to the state-based
// check (CR 704.5f); the boosts end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd65468c-d75a-4be2-8a0c-9c4e808e3dd0",
		Name:         "Shambling Strider",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}{G}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{R}{G}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Shambling Strider — +1/-1 until end of turn"),
		}},
	})
}
