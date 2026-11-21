package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embereth Veteran — Creature — Human Knight {R}, 2/1:
//
//	"{1}, Sacrifice this creature: Create a Young Hero Role token
//	 attached to another target creature. (If you control another Role on
//	 it, put that one into the graveyard. Enchanted creature has
//	 "Whenever this creature attacks, if its toughness is 3 or less, put
//	 a +1/+1 counter on it.")"
//
// "Another target creature" is any creature, yours or not; the Role is
// controlled by the Veteran's controller either way, so a Role an
// opponent put there is not replaced (CR 704.5z counts one player's
// Roles). The ability can be activated at instant speed.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c232056-9dd6-4a78-be08-3bff19daa88d",
		Name:         "Embereth Veteran",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Create a Young Hero Role token attached to another target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.CreateRoleOnFirstTarget(effects.RoleYoungHero),
		}},
	})
}
