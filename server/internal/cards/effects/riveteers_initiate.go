package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riveteers Initiate (Streets of New Capenna, #5491):
//
//	{1}{B/G}: This creature gains deathtouch until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c756d786-9aff-4d2e-9c06-6ae01d5fca2e",
		Name:         "Riveteers Initiate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B/G}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{B/G}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Riveteers Initiate — deathtouch until end of turn"),
		}},
	})
}
