package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mortarpod — Artifact — Equipment {2} (Mirrodin Besieged, S82 OLD b466,
// #5576):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black Phyrexian
//	 Germ creature token, then attach this to it.)
//	 Equipped creature gets +0/+1 and has "Sacrifice this creature: This
//	 creature deals 1 damage to any target."
//	 Equip {2}"
//
// Living weapon (the shared trigger), a layer 7c +0/+1 and a granted ability
// bundle: the creature is sacrificed to pay (CR 118.12) and the damage is
// dealt by it as it last existed (CR 608.2h), so a lifelinking or deathtouching
// creature keeps those on the damage. The grant works for whatever creature
// the Mortarpod is attached to and stops when it is unattached.
//
// No simplification.
const mortarpodGrant = "mortarpod/deal-1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "77248a76-373c-48fc-939f-aa378ab57f7c",
		Name:         "Mortarpod",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: mortarpodGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "Sacrifice this creature: This creature deals 1 damage to any target.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
				Cost:    effects.SacrificeThis(),
				Targets: effects.TargetAny(),
				Effect:  effects.OldB466SourceDealsOneDamage,
			}},
			Text: "Sacrifice this creature: This creature deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 1),
			effects.GrantAbilitiesToAttached(mortarpodGrant),
		},
		Triggered: []game.TriggeredAbility{effects.LivingWeapon("Mortarpod")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
