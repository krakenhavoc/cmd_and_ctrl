package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uktabi Faerie — Creature — Faerie {1}{G}, 1/1 (Mirage):
//
//	"Flying
//	 {3}{G}, Sacrifice this creature: Destroy target artifact."
//
// The sacrifice is a cost paid at announce (CR 602.2b); the artifact is
// checked again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "447f1cce-bbc5-473b-a838-34070f8fd2dc",
		Name:            "Uktabi Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, Sacrifice this creature: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
