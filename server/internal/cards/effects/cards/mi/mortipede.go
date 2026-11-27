package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortipede — Creature — Insect {3}{B}, 4/1 (Ravnica: City of Guilds, S82 OLD
// batch 574, #5600):
//
//	"{2}{G}: All creatures able to block this creature this turn do so."
//
// A lure on the Mortipede until end of turn (Bloodscent's requirement,
// CR 509.1c): every creature able to block it must. It binds only while the
// Mortipede is still the permanent that was activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5122d52-cebc-435b-a077-cc64376d7a86",
		Name:         "Mortipede",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: All creatures able to block this creature this turn do so.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BlockRequirementUntilEOT{
					Target: item.SourceCardID,
					Kind:   game.BlockRequirementLure,
					Label:  "Mortipede — all creatures able to block it do so",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
