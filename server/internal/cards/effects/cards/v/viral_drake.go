package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viral Drake — Creature — Phyrexian Drake {3}{U}, 1/4 (New Phyrexia, #5575):
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	 {3}{U}: Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Flying and infect (CR 702.9, CR 702.90), and Lulu's proliferate row:
// the controller chooses what to proliferate (CR 701.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b1647d1-9aec-454f-a215-e28633cb7a2d",
		Name:            "Viral Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.Do(effects.Proliferate{}),
		}},
	})
}
