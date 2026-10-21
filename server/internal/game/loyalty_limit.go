package game

import "github.com/google/uuid"

// loyalty_limit.go — CR 606.3's "only one loyalty ability per
// permanent per turn", and the printed exception to it: "You may
// activate the loyalty abilities of Urza twice each turn rather than
// only once" (Urza, Planeswalker; ADR 0145), and its class form, "You
// may activate the loyalty abilities of planeswalkers you control twice
// each turn rather than only once" (Oath of Teferi, #4184). Two sources
// of the static do not stack: the limit is two either way.
//
// Two maps, both keyed by the permanent's instance ID and both flushed
// at the turn boundary and forgotten when the object leaves the
// battlefield: LoyaltyActivatedThisTurn (the first activation, as it
// always was) and LoyaltyActivatedTwiceThisTurn (the second, which only
// a permanent with the static can make). Every reader asks
// LoyaltySpentLocked, so the activation gate, the sandbox verb, the
// legal-move enumerator and the wire's loyalty_activated flag cannot
// disagree about whether a third activation is open.
//
// On top of that limit, The Chain Veil (#4191; ADR 0145 amendment
// 2026-10-21): "For each planeswalker you control, you may activate one
// of its loyalty abilities once this turn as though none of its loyalty
// abilities have been activated this turn." It modifies the rules of
// the game, so it is the PLAYER's for the rest of the turn and covers a
// planeswalker that comes under their control after it resolves (the
// card's 2014-07-18 rulings); each resolution is one more grant, and
// they stack with each other and with the "twice" statics. Two more
// maps: LoyaltyExtraActivationsThisTurn (grants, per player) and
// LoyaltyExtraActivationsUsedThisTurn (grants each permanent has spent,
// per player). A grant is spent only by an activation made once the
// ordinary limit is used up, so "as though none have been activated"
// is one activation that ignores CR 606.3's history, whatever order
// the statics come and go in.

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
// activation left this turn (CR 606.3): its ordinary limit is used up
// (it has made one and may make only one, or it has made its two) and
// it has no Chain Veil grant of its controller's left to spend.
//
// Caller must hold g.mu (read or write).
func (g *Game) LoyaltySpentLocked(cardID uuid.UUID) bool {
	if !g.loyaltyLimitSpentLocked(cardID) {
		return false
	}
	c := findBattlefieldCard(g, cardID)
	return c == nil || !g.loyaltyExtraOpenLocked(c)
}

// loyaltyLimitSpentLocked is CR 606.3's ordinary limit alone, as
// raised to two by Urza's and Oath of Teferi's statics: the permanent
// has made one activation and may make only one, or has made two.
//
// Caller must hold g.mu (read or write).
func (g *Game) loyaltyLimitSpentLocked(cardID uuid.UUID) bool {
	if !g.LoyaltyActivatedThisTurn[cardID] {
		return false
	}
	if g.LoyaltyActivatedTwiceThisTurn[cardID] {
		return true
	}
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return true
	}
	if c.loyaltyTwiceEachTurn() {
		return false
	}
	return !c.IsPlaneswalker() || !g.yourPlaneswalkersLoyaltyTwiceLocked(c.Controller)
}

// CatalogYourPlaneswalkersLoyaltyTwiceEachTurn reports whether a
// battlefield permanent with this catalog key carries the printed static
// "You may activate the loyalty abilities of planeswalkers you control
// twice each turn rather than only once" (Oath of Teferi). Populated
// from effects.Spec.YourPlaneswalkersLoyaltyTwiceEachTurn. A nil hook
// grants nothing.
var CatalogYourPlaneswalkersLoyaltyTwiceEachTurn func(key string) bool

// yourPlaneswalkersLoyaltyTwiceLocked reports whether a permanent on the
// battlefield controlled by `player` has the Oath of Teferi static, read
// live through CatalogAbilityKey: a source that has lost its abilities,
// has left, or is phased out (then it is not on the battlefield, CR
// 702.26b) grants nothing.
//
// Caller must hold g.mu (read or write).
func (g *Game) yourPlaneswalkersLoyaltyTwiceLocked(player uuid.UUID) bool {
	if CatalogYourPlaneswalkersLoyaltyTwiceEachTurn == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogYourPlaneswalkersLoyaltyTwiceEachTurn(key) {
			return true
		}
	}
	return false
}

// loyaltyExtraOpenLocked reports whether the planeswalker c has a
// Chain Veil grant of its controller's left to spend this turn. Only
// the controller's grants count: loyalty abilities are activated by
// the permanent's controller (CR 602.2), and a grant is "you may
// activate". A permanent that is not a planeswalker has none ("for each
// planeswalker you control").
//
// Caller must hold g.mu (read or write).
func (g *Game) loyaltyExtraOpenLocked(c *Card) bool {
	n := g.LoyaltyExtraActivationsThisTurn[c.Controller]
	if n <= 0 || !c.IsPlaneswalker() {
		return false
	}
	return g.LoyaltyExtraActivationsUsedThisTurn[c.InstanceID][c.Controller] < n
}

// GrantExtraLoyaltyActivationForEffect is The Chain Veil's resolution
// (#4191): for the rest of this turn, each planeswalker `player`
// controls may activate one of its loyalty abilities once more as
// though none had been activated. One grant per call; two resolutions
// (an untap and a second activation) make two. Read live against the
// planeswalkers the player controls, so one that comes under their
// control later this turn is covered too. Expires at the turn boundary
// with the rest of the per-turn state.
//
// Caller must hold g.mu (write).
func (g *Game) GrantExtraLoyaltyActivationForEffect(player uuid.UUID) {
	if player == uuid.Nil {
		return
	}
	if g.LoyaltyExtraActivationsThisTurn == nil {
		g.LoyaltyExtraActivationsThisTurn = make(map[uuid.UUID]int)
	}
	g.LoyaltyExtraActivationsThisTurn[player]++
}

// recordLoyaltyActivationLocked notes one loyalty activation of the
// permanent: the first in LoyaltyActivatedThisTurn, a second in
// LoyaltyActivatedTwiceThisTurn. An activation made once the ordinary
// limit is used up spends one of its controller's Chain Veil grants
// (LoyaltyExtraActivationsUsedThisTurn); the gate only lets it through
// when one is left.
//
// Caller must hold g.mu.
func (g *Game) recordLoyaltyActivationLocked(cardID uuid.UUID) {
	if g.loyaltyLimitSpentLocked(cardID) {
		if c := findBattlefieldCard(g, cardID); c != nil && g.loyaltyExtraOpenLocked(c) {
			if g.LoyaltyExtraActivationsUsedThisTurn == nil {
				g.LoyaltyExtraActivationsUsedThisTurn = make(map[uuid.UUID]map[uuid.UUID]int)
			}
			used := g.LoyaltyExtraActivationsUsedThisTurn[cardID]
			if used == nil {
				used = make(map[uuid.UUID]int)
				g.LoyaltyExtraActivationsUsedThisTurn[cardID] = used
			}
			used[c.Controller]++
		}
	}
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
