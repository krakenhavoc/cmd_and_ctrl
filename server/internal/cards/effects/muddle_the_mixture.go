package effects

// Muddle the Mixture — Instant {U}{U}:
//
//	"Counter target instant or sorcery spell.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal
//	 it, put it into your hand, then shuffle. Transmute only as a
//	 sorcery.)"
//
// Transmute is the shared keyword constructor (transmute.go): an
// activated ability from the hand, paid by discarding the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0826cf7d-7ccc-459b-a92f-29dc169628f8",
		Name:         "Muddle the Mixture",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		Activated:    []ActivatedAbility{Transmute("{1}{U}{U}")},
		OnResolve:    counterTheTargetSpell,
	})
}
