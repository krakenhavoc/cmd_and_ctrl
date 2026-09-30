package effects

// Death Rattle — Instant {5}{B}:
//
//	"Delve. Destroy target nongreen creature. It can't be regenerated."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae327976-1026-4318-8846-f6b0cac373a7",
		Name:         "Death Rattle",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets:      TargetCreature("target nongreen creature", Not(OfColor("G"))),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
