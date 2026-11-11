package game

import "github.com/google/uuid"

// multiply_tokens.go — a token doubler a spell or ability makes for the
// rest of the turn (Kaya, Geist Hunter, #5368): "until end of turn, if one
// or more tokens would be created under your control, twice that many of
// those tokens are created instead" (CR 614.1a, 701.7b).

// ModMultiplyTokens multiplies the tokens one create-tokens instruction
// would make under the record's player's control by Amount (2 for
// "twice"). The same replacement Parallel Lives is, as a turn-scoped
// record: one event per instruction, ×2 twice being ×4 (CR 616.1). Reads
// Amount and Player. Scope ScopeGame. An on-disk identity: never renamed,
// never reused.
const ModMultiplyTokens ModKind = "multiplyTokens"

// multiplyTokensAppliesLocked is ModMultiplyTokens's AppliesTo.
func multiplyTokensAppliesLocked(m Mod, ev *ReplacementEvent) bool {
	return ev.Kind == RepEventCreateTokens && ev.TokenCount() > 0 && m.Amount >= 2 &&
		m.Player != uuid.Nil && ev.TokenController == m.Player
}

// MultiplyTokensThisTurnForEffect registers a ModMultiplyTokens record for
// `player` until end of turn. Reports whether a record was written (not
// for a factor below 2 or no player).
//
// Caller must hold g.mu (write).
func (g *Game) MultiplyTokensThisTurnForEffect(sourceID, player uuid.UUID, factor int, label string) bool {
	if player == uuid.Nil || factor < 2 {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModMultiplyTokens, Amount: factor, Player: player}}, g.UntilEndOfTurnDuration(), label)
}
