package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roofstalker Wight — Creature — Zombie {1}{B}, 2/1 (Ravnica: Clue Edition):
//
//	"{1}{U}: This creature gains flying until end of turn."
//
// Afterthought Sentry's self grant on its own mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2f2fd39-4a21-489b-893e-4258f8d2aa63",
		Name:         "Roofstalker Wight",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"flying"},
					Label:    "Roofstalker Wight — flying until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
