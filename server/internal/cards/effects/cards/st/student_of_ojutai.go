package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Student of Ojutai — Creature — Human Monk {3}{W}, 2/4 (Dragons of Tarkir, #5551):
//
//	"Whenever you cast a noncreature spell, you gain 2 life."
//
// Monastery Mentor's cast trigger with a life-gain body.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "0adcf8d2-6eb9-4679-8ad5-f28e2f4aba39",
		Name:         "Student of Ojutai",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Student of Ojutai — you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
