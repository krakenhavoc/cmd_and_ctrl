package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voldaren Thrillseeker — Creature — Vampire Warrior {2}{R}, 1/1 (March of
// the Machine):
//
//	"Backup 2 (When this creature enters, put two +1/+1 counters on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 {1}, Sacrifice this creature: It deals damage equal to its power to any
//	 target."
//
// Heartfire Immolator's ability on any target
// (sacrificedSourceDealsItsPowerToTheTarget): the sacrifice is a cost, so
// "its power" is its power as it last existed on the battlefield
// (CR 113.7a), counters included. Backup grants it as a bundle.
//
// No simplification.
const voldarenThrillseekerGrant = "voldaren-thrillseeker/sacrifice-damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff",
		Name:         "Voldaren Thrillseeker",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       voldarenThrillseekerGrant,
			Activated: []effects.ActivatedAbility{voldarenThrillseekerSacrifice()},
			Text:      "{1}, Sacrifice this creature: It deals damage equal to its power to any target.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Voldaren Thrillseeker", 2, voldarenThrillseekerGrant),
		},
		Activated: []effects.ActivatedAbility{voldarenThrillseekerSacrifice()},
	})
}

func voldarenThrillseekerSacrifice() effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label:   "{1}, Sacrifice this creature: It deals damage equal to its power to any target.",
		Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
		Targets: effects.TargetAny(),
		Effect:  effects.SacrificedSourceDealsItsPowerToTheTarget,
	}
}
