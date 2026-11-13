package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honor Guard — Creature — Human Soldier {W}, 1/1 (Stronghold, #5627):
//
//	"{W}: This creature gets +0/+1 until end of turn."
//
// Pearl Dragon's +0/+1 at {W}, with the shared thisGetsUntilEndOfTurn
// body: no tap in the cost, so it can be activated any number of times,
// summoning sick or not, and the boost ends in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ba82d8e-1857-40a9-8b42-5d8348c68859",
		Name:         "Honor Guard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGetsUntilEndOfTurn(0, 1, "Honor Guard — +0/+1 until end of turn"),
		}},
	})
}
