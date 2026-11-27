package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcish Settlers — Creature — Orc {1}{R}, 1/1 (Weatherlight):
//
//	"{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands."
//
// Exactly X targets, X being the announced value (CR 107.3); X may be 0, which
// destroys nothing. Targets that are gone or illegal on resolution are skipped
// (CR 608.2b) and the rest are destroyed together.
//
// No simplification.
func init() {
	targets := effects.TargetPermanent("X target lands", effects.Land())
	targets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "8b8d4ae0-0d18-4f06-a883-0edff494d06b",
		Name:         "Orcish Settlers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{X}{X}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			},
		}},
	})
}
