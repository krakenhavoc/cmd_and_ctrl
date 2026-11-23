package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Augur — Creature — Bird Wizard {3}{U}, 2/2 (Future Sight, #5594):
//
//	"Flying
//	 Sacrifice this creature: Return up to two target creatures to their
//	 owners' hands. Activate only during your upkeep."
//
// An upkeep-only sacrifice (DuringYourUpkeep is the step AND your turn)
// whose body is the shared bounce of every still-legal target; none
// chosen is a legal announcement. It can send an attacker-to-be home or
// save a creature of yours, so it declares both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e68d1b86-3996-4bbf-9086-a76a7f9ce2a4",
		Name:            "Aven Augur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice this creature: Return up to two target creatures to their owners' hands. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove | game.AnswerSurvive},
			Cost:      effects.SacrificeThis(),
			Condition: effects.DuringYourUpkeep(),
			Targets:   effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			Effect:    effects.BounceEachLegalTarget,
		}},
	})
}
