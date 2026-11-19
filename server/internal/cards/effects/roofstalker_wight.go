package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roofstalker Wight — Creature — Zombie {1}{B}, 2/1 (Ravnica: Clue Edition):
//
//	"{1}{U}: This creature gains flying until end of turn."
//
// Afterthought Sentry's self grant on its own mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2f2fd39-4a21-489b-893e-4258f8d2aa63",
		Name:         "Roofstalker Wight",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"flying"},
					Label:    "Roofstalker Wight — flying until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
