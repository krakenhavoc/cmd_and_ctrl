package game

import (
	"slices"

	"github.com/google/uuid"
)

// sacrifice_grantor_cost.go — AbilityCost.SacrificeGrantor: a granted
// ability whose cost sacrifices the permanent that granted it (Deconstruction
// Hammer). It reads the grantor off the row's origin like ExileGrantor and
// joins it to the payment's sacrifice list, so the one sacrifice route pays
// it (CR 701.21a, CR 118.3).

// addSacrificeGrantorLocked appends the row's grantor to the sacrifices when
// the cost names it. ErrInvalidParam when the row has no grantor, the grantor
// left the battlefield or is already a pick; ErrCardCallerMismatch when the
// payer doesn't control it (you can only sacrifice what you control);
// ErrCantBeSacrificed when it can't be. Caller must hold g.mu.
func (g *Game) addSacrificeGrantorLocked(playerID uuid.UUID, cost AbilityCost, origin AbilityOrigin, sacrifices []uuid.UUID) ([]uuid.UUID, error) {
	if !cost.SacrificeGrantor {
		return sacrifices, nil
	}
	if !origin.Granted() || origin.GrantedBy == uuid.Nil || g.Battlefield == nil || !g.Battlefield.Contains(origin.GrantedBy) {
		return nil, ErrInvalidParam
	}
	c := findBattlefieldCard(g, origin.GrantedBy)
	if c == nil {
		return nil, ErrCardNotFound
	}
	if c.Controller != playerID {
		return nil, ErrCardCallerMismatch
	}
	if g.permanentCantBeSacrificedLocked(c) {
		return nil, ErrCantBeSacrificed
	}
	if slices.Contains(sacrifices, origin.GrantedBy) {
		return nil, ErrInvalidParam
	}
	return append(sacrifices, origin.GrantedBy), nil
}
