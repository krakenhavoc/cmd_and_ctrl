package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caregiver — Creature — Human Cleric {W}, 1/1 (Ravnica: City of Guilds,
// #5600):
//
//	"{W}, Sacrifice a creature: Prevent the next 1 damage that would be
//	 dealt to any target this turn."
//
// Benevolent Ancestor's shield with a creature sacrificed as the cost,
// which may be Caregiver itself: the cost is paid before the ability
// resolves, and the shield does not need its source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7661484b-1faa-4ed7-b77b-fc1e625d23e3",
		Name:         "Caregiver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice a creature: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Caregiver"),
		}},
	})
}
