package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mold Folk — {1}{B} Creature — Fungus Warrior 1/1
//
//	"Lifelink
//	 Mold Harvest — {1}, Sacrifice another creature or an artifact: Put a
//	 +1/+1 counter on this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d0ec32d9-1d4e-4b6f-bc95-addbb484b0d3",
		Name:            "Mold Folk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Mold Harvest — {1}, Sacrifice another creature or an artifact: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost: effects.Plus(effects.ManaCost("{1}"),
				effects.SacrificeAnotherN(1, "another creature or an artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Effect: effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
