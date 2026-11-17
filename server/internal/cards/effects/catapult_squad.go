package effects

// Catapult Squad — Creature — Human Soldier {1}{W}, 2/1 (Onslaught):
//
//	"Tap two untapped Soldiers you control: This creature deals 2 damage
//	 to target attacking or blocking creature."
//
// Shieldmage Elder's tap-two cost over Soldiers (tapTwoUntapped), the
// Squad among them if its controller likes, and with no {T} symbol a
// Soldier that arrived this turn may pay it (CR 302.6). The damage is
// the Squad's own, so it is dealt even when the Squad was one of the
// two tapped. The target is any attacking or blocking creature, on
// either side.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1ff7a67-1add-4de1-bc25-62f63626c43a",
		Name:         "Catapult Squad",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped Soldiers you control: This creature deals 2 damage to target attacking or blocking creature.",
			Cost:    tapTwoUntapped("two untapped Soldiers you control", "Soldier"),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
