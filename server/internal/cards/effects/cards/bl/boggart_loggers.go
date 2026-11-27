package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Loggers — Creature — Goblin Rogue {2}{B}, 2/1 (Lorwyn, S82 OLD b557):
//
//	"Forestwalk
//	 {2}{B}, Sacrifice this creature: Destroy target Treefolk or Forest."
//
// The target is a Treefolk creature or a Forest land (a permanent with either
// subtype). Sacrificing is a cost, so the ability resolves even when the
// Loggers are gone (CR 112.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8fac28d-adf4-47e9-bdd3-8402cbf0f228",
		Name:            "Boggart Loggers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this creature: Destroy target Treefolk or Forest.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target Treefolk or Forest", effects.Or(effects.HasSubtype("Treefolk"), effects.HasSubtype("Forest"))),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
