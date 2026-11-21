package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Albino Troll — Creature — Troll, {1}{G}, 3/3:
//
//	"Echo {1}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {1}{G}: Regenerate this creature."
//
// Regeneration is Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02079e18-ac14-4255-95c4-741eb76c4799",
		Name:         "Albino Troll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Albino Troll", "{1}{G}")},
	})
}
