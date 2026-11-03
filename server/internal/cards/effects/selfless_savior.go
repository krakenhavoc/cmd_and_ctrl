package effects

// Selfless Savior — Creature — Dog {W}, 1/1:
//
//	"Sacrifice this creature: Another target creature you control gains
//	 indestructible until end of turn."
//
// The sacrifice is paid at announce. "Another" is object identity
// (CR 109.1, effects.Another), so the Savior can't name itself; the
// target must still be a creature you control as the ability resolves
// (CR 608.2b), or nothing happens.
//
// No simplification.
func init() {
	const label = "Selfless Savior — another creature you control gains indestructible until end of turn"
	Register(Spec{
		OracleID:     "49dc3434-2501-4788-84f2-93e55730f67c",
		Name:         "Selfless Savior",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Selfless Savior: Another target creature you control gains indestructible until end of turn.",
			Cost:    SacrificeThis(),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  firstTargetGainsUntilEndOfTurn(0, label, "indestructible"),
		}},
	})
}
