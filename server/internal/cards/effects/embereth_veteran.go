package effects

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
	Register(Spec{
		OracleID:     "2c232056-9dd6-4a78-be08-3bff19daa88d",
		Name:         "Embereth Veteran",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Create a Young Hero Role token attached to another target creature.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  createRoleOnFirstTarget(RoleYoungHero),
		}},
	})
}
