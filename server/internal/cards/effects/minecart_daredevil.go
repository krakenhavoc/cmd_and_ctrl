package effects

// Minecart Daredevil // Ride the Rails — Creature — Dwarf Knight {2}{R},
// 4/2 // Instant — Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Minecart Daredevil — no rules text.
//	Ride the Rails — "Target creature gets +2/+1 until end of turn. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"), as Foulmire Knight's do: the creature half has no text, but
// the card prints rules on its other half, so the creature registers to
// say it is complete. Ride the Rails is Artful Maneuver's one-target pump
// at +2/+1; the Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const minecartDaredevilOracleID = "df555f60-2164-49dc-be67-00aca9eb0d45"

func init() {
	Register(Spec{
		OracleID:     minecartDaredevilOracleID,
		Name:         "Minecart Daredevil",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     minecartDaredevilOracleID + "#1",
		Name:         "Ride the Rails",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 1)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 1, "Ride the Rails — +2/+1"),
	})
}
