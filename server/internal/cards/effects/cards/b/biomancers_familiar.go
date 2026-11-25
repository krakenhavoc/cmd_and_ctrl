package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biomancer's Familiar — Creature — Mutant {G}{U}, 2/2 (Ravnica Allegiance,
// #5521):
//
//	"Activated abilities of creatures you control cost {2} less to activate.
//	 This effect can't reduce the mana in that cost to less than one mana.
//	 {T}: The next time target creature adapts this turn, it adapts as though
//	 it had no +1/+1 counters on it."
//
// The reduction is Training Grounds' cost modifier. The second ability
// records a one-shot on the target (#6239) that the next adapt of that
// creature spends, whatever counters it has; it ends at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "923d44a9-774d-459c-9cec-1d5b081b51f0",
		Name:         "Biomancer's Familiar",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.KeepsOneMana(effects.ActivationCostsLess(2,
				"Activated abilities of creatures you control cost {2} less to activate. This effect can't reduce the mana in that cost to less than one mana.",
				effects.OfACreatureYouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: The next time target creature adapts this turn, it adapts as though it had no +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NextAdaptIgnoresCounters("Biomancer's Familiar"),
		}},
	})
}
