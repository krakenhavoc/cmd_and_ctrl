package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff",
		Name:         "Voldaren Thrillseeker",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       voldarenThrillseekerGrant,
			Activated: []ActivatedAbility{VoldarenThrillseekerSacrifice()},
			Text:      "{1}, Sacrifice this creature: It deals damage equal to its power to any target.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Voldaren Thrillseeker", 2, voldarenThrillseekerGrant),
		},
		Activated: []ActivatedAbility{VoldarenThrillseekerSacrifice()},
	})
}

func VoldarenThrillseekerSacrifice() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{1}, Sacrifice this creature: It deals damage equal to its power to any target.",
		Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
		Targets: TargetAny(),
		Effect:  SacrificedSourceDealsItsPowerToTheTarget,
	}
}
