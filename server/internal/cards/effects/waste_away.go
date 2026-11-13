package effects

// Waste Away — Instant {4}{B} (Torment, #5612):
//
//	"As an additional cost to cast this spell, discard a card.
//	 Target creature gets -5/-5 until end of turn."
//
// The discard is DiscardCost(1), Big Score's: paid as the spell is cast,
// with the spell already on the stack (CR 601.2h), so it can't be cast
// from an otherwise empty hand, and a countered Waste Away leaves the
// card discarded. The shrink is the shared boostTheTargetUntilEOT body
// at -5/-5, ending at cleanup (CR 514.2), declared as a harm so it is
// aimed at an opponent's creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "d7f25f7f-3fa2-47ba-be33-ba0b08f26c94",
		Name:           "Waste Away",
		Completeness:   CompletenessFull,
		Purpose:        ForTargets(HarmUntilEOT(0, -5, -5)),
		AdditionalCost: DiscardCost(1),
		Targets:        TargetCreature("target creature"),
		OnResolve:      boostTheTargetUntilEOT(-5, -5, "Waste Away — -5/-5"),
	})
}
