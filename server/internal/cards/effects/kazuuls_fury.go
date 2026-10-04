package effects

// Kazuul's Fury // Kazuul's Cliffs — modal double-faced card. This file
// is the FRONT face, Instant {2}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Kazuul's Fury deals damage equal to the sacrificed creature's
//	 power to any target."
//
// The back face, Kazuul's Cliffs, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// Fling at one more mana, with a land on the back. The sacrificed
// creature's last-known power is the damage (the 2020-09-25 ruling, CR
// 608.2h), read off the payment record (ADR 0113 §1); negative power
// deals none (CR 107.1b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "f8410804-632b-4f18-9a73-6dccc7e4582d",
		Name:           "Kazuul's Fury",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetAny(),
		OnResolve:      damageEqualToSacrificedPower,
	})
}
