package effects

// Onakke Javelineer — Creature — Ogre Spirit {4}{R}, 5/4 (March of the
// Machine):
//
//	"Reach
//	 {T}: This creature deals 2 damage to target player or battle."
//
// Damage to a battle removes that many defense counters (CR 120.3h), done
// by the engine's damage path. The {T} keeps it off the turn it arrives
// (CR 602.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e834b281-279f-4df3-9c0e-764502b809c5",
		Name:            "Onakke Javelineer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to target player or battle.",
			Cost:    TapCost(),
			Targets: s82MomB12PlayerOrBattle(),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
