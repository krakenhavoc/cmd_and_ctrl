package effects

// Dizzy Spell — Instant {U} (Ravnica: City of Guilds, #5600):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// Shrink's layer 7c pump at -3/-0 with the transmute constructor
// (transmute.go, CR 702.53a), finding a mana value 1 card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7de04284-dcaa-4b63-a734-fcfe2698f83c",
		Name:         "Dizzy Spell",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -3, 0)),
		OnResolve:    boostTheTargetUntilEOT(-3, 0, "Dizzy Spell — -3/-0 until end of turn"),
		Activated:    []ActivatedAbility{Transmute("{1}{U}{U}")},
	})
}
