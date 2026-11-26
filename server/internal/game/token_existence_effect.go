package game

// token_existence_effect.go — CR 704.5d for an effect that must not see a
// token it has just put into a library.

// TokensCeaseToExistForEffect removes every token that is in a zone other
// than the battlefield (CR 704.5d, CR 111.7) at once, rather than at the next
// state-based check. The Great Aurora needs it: its tokens are shuffled into
// a library, and the cards it then draws must not include one, since a token
// that has left the battlefield is gone as soon as the rules look. It
// reports whether anything was removed.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) TokensCeaseToExistForEffect() bool {
	return g.tokenCeaseToExistSBALocked()
}
