package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argothian Sprite — Creature — Faerie {1}{G}, 2/2 (The Brothers' War, #5482):
//
//	"This creature can't be blocked by artifact creatures.
//	 {7}: Put two +1/+1 counters on this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0dbcc5d8-9533-48fd-a69d-bbc6c72aaac5",
		Name:         "Argothian Sprite",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.Artifact()), "artifact creatures"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{7}"),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
