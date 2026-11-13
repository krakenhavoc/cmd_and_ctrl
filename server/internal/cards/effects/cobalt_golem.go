package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cobalt Golem — Artifact Creature — Golem {4}, 2/3 (Mirrodin, #5606):
//
//	"{1}{U}: This creature gains flying until end of turn."
//
// Afterthought Sentry's self grant at {1}{U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e50af57-e49d-4b94-a370-44846d9f6b33",
		Name:         "Cobalt Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
					Label: "Cobalt Golem — flying"}.Apply(NewContext(g, item))
			},
		}},
	})
}
