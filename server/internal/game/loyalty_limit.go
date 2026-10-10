package game

import "github.com/google/uuid"

// loyalty_limit.go — CR 606.3's "only one loyalty ability per
// permanent per turn", and the printed exception to it: "You may
// activate the loyalty abilities of Urza twice each turn rather than
// only once" (Urza, Planeswalker; ADR 0145).
//
// Two maps, both keyed by the permanent's instance ID and both flushed
// at the turn boundary and forgotten when the object leaves the
// battlefield: LoyaltyActivatedThisTurn (the first activation, as it
// always was) and LoyaltyActivatedTwiceThisTurn (the second, which only
// a permanent with the static can make). Every reader asks
// LoyaltySpentLocked, so the activation gate, the sandbox verb, the
// legal-move enumerator and the wire's loyalty_activated flag cannot
// disagree about whether a third activation is open.

// CatalogLoyaltyTwiceEachTurn reports whether a battlefield permanent
// with this catalog key may activate its loyalty abilities twice each
// turn. Populated from effects.Spec.LoyaltyTwiceEachTurn. A nil hook
// grants nothing.
var CatalogLoyaltyTwiceEachTurn func(key string) bool

// loyaltyTwiceEachTurn answers the static for one permanent, through
// CatalogAbilityKey: a permanent that has lost its abilities is held to
// one activation (CR 613.1f).
func (c *Card) loyaltyTwiceEachTurn() bool {
	if CatalogLoyaltyTwiceEachTurn == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	return key != "" && CatalogLoyaltyTwiceEachTurn(key)
}

// LoyaltySpentLocked reports whether the permanent has no loyalty
// activation left this turn (CR 606.3): it has made one and may make
// only one, or it has made its two.
//
// Caller must hold g.mu (read or write).
func (g *Game) LoyaltySpentLocked(cardID uuid.UUID) bool {
	if !g.LoyaltyActivatedThisTurn[cardID] {
		return false
	}
	if g.LoyaltyActivatedTwiceThisTurn[cardID] {
		return true
	}
	c := findBattlefieldCard(g, cardID)
	return c == nil || !c.loyaltyTwiceEachTurn()
}

// recordLoyaltyActivationLocked notes one loyalty activation of the
// permanent: the first in LoyaltyActivatedThisTurn, a second in
// LoyaltyActivatedTwiceThisTurn.
//
// Caller must hold g.mu.
func (g *Game) recordLoyaltyActivationLocked(cardID uuid.UUID) {
	if g.LoyaltyActivatedThisTurn[cardID] {
		if g.LoyaltyActivatedTwiceThisTurn == nil {
			g.LoyaltyActivatedTwiceThisTurn = make(map[uuid.UUID]bool)
		}
		g.LoyaltyActivatedTwiceThisTurn[cardID] = true
		return
	}
	if g.LoyaltyActivatedThisTurn == nil {
		g.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool)
	}
	g.LoyaltyActivatedThisTurn[cardID] = true
}
