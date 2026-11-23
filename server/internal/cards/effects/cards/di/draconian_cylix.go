package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draconian Cylix — Artifact {3}:
//
//	"{2}, {T}, Discard a card at random: Regenerate target creature."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// The shield is CR 701.19a regeneration on the target as the ability
// resolves (CR 608.2b rechecks it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0880b36a-6141-4955-b532-cf88daca0869",
		Name:         "Draconian Cylix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Discard a card at random: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
