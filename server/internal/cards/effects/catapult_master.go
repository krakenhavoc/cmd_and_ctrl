package effects

// Catapult Master — Creature — Human Soldier {3}{W}{W}, 3/3 (Onslaught):
//
//	"Tap five untapped Soldiers you control: Exile target creature."
//
// Ancestor's Prophet's tap-five cost over Soldiers (tapFiveUntapped):
// the Master may be one of the five, and with no {T} symbol in the cost
// a Soldier that arrived this turn may pay it (CR 302.6). The exile is
// Hanged Executioner's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02da2145-0edb-44be-8622-360182a8c25a",
		Name:         "Catapult Master",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap five untapped Soldiers you control: Exile target creature.",
			Cost:    tapFiveUntapped("five untapped Soldiers you control", "Soldier"),
			Targets: TargetCreature("target creature"),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
