package effects

// Tor Wauki — Legendary Creature — Human Archer {2}{B}{B}{R}, 3/3
// (Legends, #5639):
//
//	"{T}: Tor Wauki deals 2 damage to target attacking or blocking
//	 creature."
//
// The target clause is AttackingOrBlocking (Eiganjo, Seat of the
// Empire's), which reads the combat designations, so the ability has a
// target only during a combat with attackers or blockers. The damage is
// the shared DealDamageToTheTarget body, dealt by Tor Wauki. The tap
// means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9439dfa0-daf8-4dac-af32-755fc2984126",
		Name:         "Tor Wauki",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tor Wauki deals 2 damage to target attacking or blocking creature.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
