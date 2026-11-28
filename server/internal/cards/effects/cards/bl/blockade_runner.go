package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blockade Runner — Creature — Merfolk {3}{U}, 2/2 (Mercadian Masques, S82 OLD
// b682, #5619):
//
//	"{U}: This creature can't be blocked this turn."
//
// Agent of Horizons' grant (RestrictUntilEOT, CantBeBlocked) for {U}: nothing
// if the source has left or come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f97df8f4-a7ef-45f9-9563-b76b4d2b03f5",
		Name:         "Blockade Runner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Blockade Runner — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
