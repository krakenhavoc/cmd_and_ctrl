package effects

// Pyre Charger — Creature — Elemental Warrior {R}{R}, 1/1 (Shadowmoor,
// #5591):
//
//	"Haste
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Haste is a printed keyword. thisGetsRow, Kranioceros' self pump: an
// activation with no {T}, so it can be activated any number of times. A
// hybrid symbol is paid with either colour (CR 107.4e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "50bc464e-1c8d-455c-a870-ddbbc4618343",
		Name:            "Pyre Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []ActivatedAbility{thisGetsRow("{R}", 1, 0, "Pyre Charger")},
	})
}
