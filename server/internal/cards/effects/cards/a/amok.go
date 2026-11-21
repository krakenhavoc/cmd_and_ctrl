package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amok — Enchantment {1}{R}:
//
//	"{1}, Discard a card at random: Put a +1/+1 counter on target creature."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ddbbe65-f928-4b8e-8c4c-a9ce82e2594d",
		Name:         "Amok",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Discard a card at random: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
		}},
	})
}
