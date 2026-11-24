package game

import "github.com/google/uuid"

// loyalty_limit_more.go — #6100: "you may activate Comet's loyalty ability
// two more times this turn" (Comet, Stellar Pup). One permanent's grant of
// N more loyalty activations this turn, beyond CR 606.3's limit as any
// "twice" static or effect has raised it. Each resolution adds N more, so
// a second 6 on the same turn grants two more again. A grant is spent
// only by an activation made once the ordinary limit is used up
// (recordLoyaltyActivationLocked), before any Chain Veil grant of the
// controller's, since it can pay for nothing else. It stays with the
// object: it is forgotten when the permanent leaves the battlefield (CR
// 400.7) and flushed at the turn boundary. The other timing rules of CR
// 606.3 (main phase, empty stack, priority) still apply.

// GrantThisPermanentMoreLoyaltyActivationsForEffect is the resolution of
// "you may activate [this permanent]'s loyalty ability N more times this
// turn". A permanent no longer on the battlefield gets nothing.
//
// Caller must hold g.mu (write).
func (g *Game) GrantThisPermanentMoreLoyaltyActivationsForEffect(cardID uuid.UUID, n int) {
	if cardID == uuid.Nil || n <= 0 || findBattlefieldCard(g, cardID) == nil {
		return
	}
	if g.LoyaltyPermanentExtraThisTurn == nil {
		g.LoyaltyPermanentExtraThisTurn = make(map[uuid.UUID]int)
	}
	g.LoyaltyPermanentExtraThisTurn[cardID] += n
}

// loyaltyPermanentExtraOpenLocked reports whether the permanent has one of
// its own extra activations left to spend this turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) loyaltyPermanentExtraOpenLocked(cardID uuid.UUID) bool {
	return g.LoyaltyPermanentExtraUsedThisTurn[cardID] < g.LoyaltyPermanentExtraThisTurn[cardID]
}
