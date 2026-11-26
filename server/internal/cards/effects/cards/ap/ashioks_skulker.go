package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok's Skulker — Creature — Nightmare {4}{U}, 3/5 (War of the Spark, #5520):
//
//	"{3}{U}: This creature can't be blocked this turn."
//
// Frostpeak Yeti's grant: a restriction on the Skulker until end of turn, only
// if it is still the same permanent as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffaa2776-f61d-4eaa-8949-a0407f85214f",
		Name:         "Ashiok's Skulker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Ashiok's Skulker — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
