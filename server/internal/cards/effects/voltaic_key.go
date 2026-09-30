package effects

// Voltaic Key — Artifact {1} (EDHREC rank 1752):
//
//	"{1}, {T}: Untap target artifact."
//
// The one-mana untapper. A mana-plus-tap activated ability with an
// artifact target clause; the untap is the primitive. "Target
// artifact" admits the Key itself, as printed — pointless, but the
// rules allow it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09aeea91-b1dc-443f-a509-4758f052c0a7",
		Name:         "Voltaic Key",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target artifact.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  untapFirstLegalTarget,
		}},
	})
}
