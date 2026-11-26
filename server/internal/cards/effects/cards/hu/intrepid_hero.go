package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intrepid Hero — Creature — Human Soldier {2}{W}, 1/1 (Magic 2013, S82 OLD, #5568):
//
//	"{T}: Destroy target creature with power 4 or greater."
//
// A tap ability (so not while summoning sick, CR 302.6) whose target must have
// power 4 or greater as it is chosen and again as it resolves (CR 608.2b), read
// through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63c1dfcc-9d47-4dee-b4e6-1f50e716cf9e",
		Name:         "Intrepid Hero",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature with power 4 or greater.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
