package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne of the High City — Land:
//
//	"{T}: Add {C}.
//	 {4}, {T}, Sacrifice this land: You become the monarch."
//
// The colourless monarch land (#1722). The second line is an ordinary
// CR 602 activated ability — it uses the stack, at instant speed — whose
// effect is BecomeTheMonarch. Its costs are paid at announce, so the
// land is gone before the ability resolves and a Stifle leaves you with
// neither. Activating it while you already hold the crown changes
// nothing: you are not "becoming" the monarch (CR 725.3), so nothing
// that watches for it triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9684447a-5955-4bc7-8ad0-8bb8b316873b",
		Name:         "Throne of the High City",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this land: You become the monarch",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.BecomeTheMonarch{}),
		}},
	})
}
