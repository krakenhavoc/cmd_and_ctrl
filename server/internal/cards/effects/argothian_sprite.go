package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argothian Sprite — Creature — Faerie {1}{G}, 2/2 (The Brothers' War, #5482):
//
//	"This creature can't be blocked by artifact creatures.
//	 {7}: Put two +1/+1 counters on this creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dbcc5d8-9533-48fd-a69d-bbc6c72aaac5",
		Name:         "Argothian Sprite",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), Artifact()), "artifact creatures"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{7}: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{7}"),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
