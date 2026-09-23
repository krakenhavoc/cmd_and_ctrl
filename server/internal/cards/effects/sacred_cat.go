package effects

// Sacred Cat — Creature — Cat {W}, 1/1:
//
//	"Lifelink
//	 Embalm {W} ({W}, Exile this card from your graveyard: Create a
//	 token that's a copy of it, except it's a white Zombie Cat with
//	 no mana cost. Embalm only as a sorcery.)"
//
// #1221's embalm proof. Lifelink is a printed keyword the engine
// already runs off the imported card, so the Spec is the keyword
// ability alone.
//
// The reminder text is the clearest statement of CR 707.9b's ADD
// form in the whole keyword: "a white Zombie CAT" spells the retained
// creature type out, which is why embalm.go's exception builds its
// type line with addedSubtypeTypeLine rather than with
// retypedTypeLine — the set form would create a plain Zombie and
// quietly break every Cat tribal payoff.
//
// The token copies the card as it sits IN EXILE, which is where the
// activation cost has just put it. Its lifelink rides along, because
// a printed keyword is a copiable value (CR 707.2).
func init() {
	Register(Spec{
		OracleID:     "d85ea576-a794-44bf-b405-1f1c49477409",
		Name:         "Sacred Cat",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Embalm("{W}")},
	})
}
