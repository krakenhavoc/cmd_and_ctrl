package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Albino Troll — Creature — Troll, {1}{G}, 3/3:
//
//	"Echo {1}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {1}{G}: Regenerate this creature."
//
// Regeneration is Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02079e18-ac14-4255-95c4-741eb76c4799",
		Name:         "Albino Troll",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Albino Troll", "{1}{G}")},
	})
}
