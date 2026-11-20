package effects

// Ichor Slick — Sorcery {2}{B} (Future Sight, #5594):
//
//	"Target creature gets -3/-3 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)
//	 Madness {3}{B} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Last Gasp's shrink at sorcery speed, with the shared cycling row and a
// madness cost (CR 702.35a): the exile on discard and the cast it offers
// are the keyword's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a36fc2e-bb02-4222-a335-4c80e380e690",
		Name:         "Ichor Slick",
		Completeness: CompletenessFull,
		Madness:      "{3}{B}",
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:      TargetCreature("target creature"),
		Activated:    []ActivatedAbility{Cycling("{2}")},
		OnResolve:    boostTheTargetUntilEOT(-3, -3, "Ichor Slick — -3/-3 until end of turn"),
	})
}
