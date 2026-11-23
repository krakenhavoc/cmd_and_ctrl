package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruins Recluse — Creature — Spider 1/1 {1}{G} (March of the Machine, S82):
//
//	"Reach, deathtouch
//	 {3}{G}: Put a +1/+1 counter on this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d99efecd-1419-434b-9de6-5fbb04f57dda",
		Name:            "Ruins Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
