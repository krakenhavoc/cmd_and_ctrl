package game

import "github.com/google/uuid"

// redirect_tokens.go — a token redirect a creature's ETB makes for the
// rest of the turn (Crafty Cutpurse, C21): "each token that would be
// created under an opponent's control this turn is created under your
// control instead" (CR 614.1a, 701.7b).

// ModRedirectTokens makes every token an opponent of the record's player
// would create this turn enter under the record's player's control. The
// create-tokens event's TokenController is rewritten, so each
// token is minted and enters under the new controller. Reads Player (the
// player the tokens are redirected to). Scope ScopeGame. An on-disk
// identity: never renamed, never reused.
const ModRedirectTokens ModKind = "redirectTokens"

// redirectTokensAppliesLocked is ModRedirectTokens's AppliesTo: any
// creation under another player's control (every other player is an
// opponent at a free-for-all table).
func redirectTokensAppliesLocked(m Mod, ev *ReplacementEvent) bool {
	return ev.Kind == RepEventCreateTokens && ev.TokenCount() > 0 &&
		m.Player != uuid.Nil && ev.TokenController != uuid.Nil && ev.TokenController != m.Player
}

// RedirectOpponentTokensThisTurnForEffect registers a ModRedirectTokens
// record for `player` until end of turn. Reports whether a record was
// written (not for no player).
//
// Caller must hold g.mu (write).
func (g *Game) RedirectOpponentTokensThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModRedirectTokens, Player: player}}, g.UntilEndOfTurnDuration(), label)
}
