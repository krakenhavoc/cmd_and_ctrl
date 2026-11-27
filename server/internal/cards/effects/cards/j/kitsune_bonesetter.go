package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitsune Bonesetter — Creature — Fox Cleric {2}{W}, 0/1 (Saviors of
// Kamigawa, #5601):
//
//	"{T}: Prevent the next 3 damage that would be dealt to target
//	 creature this turn. Activate only if you have more cards in hand than
//	 each opponent."
//
// The charged shield of Kitsune Healer (CR 615.7) on the target creature.
// The condition is an activation gate (CR 602.5b), judged when the ability
// is activated against every opponent still in the game; once it is on the
// stack, a change in hand sizes does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a64684bc-61b2-4261-8749-fc8cae9e2fd5",
		Name:         "Kitsune Bonesetter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Prevent the next 3 damage that would be dealt to target creature this turn. Activate only if you have more cards in hand than each opponent.",
			Purpose:   game.Purpose{Answers: game.AnswerPrevent},
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.OldB589MoreCardsInHandThanEachOpponent(),
			Effect:    effects.PreventNextDamageToTheTarget("Kitsune Bonesetter", 3),
		}},
	})
}
