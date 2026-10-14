package game

import "github.com/google/uuid"

// "Assigns no combat damage this turn" (ADR 0187 §4, #3569; CR 510.1a):
// Carrion Rats and Carrion Wurm, and the 24 cards waiting on #3531.
//
// It is a rule about how damage is assigned, not a prevention shield
// (CR 615). A creature that would assign 0 or less "doesn't assign combat
// damage at all" (CR 510.1a), so under "damage can't be prevented" it
// still assigns nothing, and nothing that watches for assigned or
// prevented damage fires. The gate has the ModCantBeRegenerated shape: a
// scoped effect pinned to the object until cleanup, read at one place.

// permanentAssignsNoCombatDamageLocked reports whether a live
// ModAssignsNoCombatDamage record covers the permanent `c`: pinned to that
// object (CR 400.7), until cleanup (CR 514.2). The combat damage pass asks
// it where it builds each combatant's amount
// (assignAndDealCombatDamageLocked), ahead of combatDamagePower, so
// DamageByToughness and DamageAsPositivePower cannot lift the 0.
//
// Caller must hold g.mu.
func (g *Game) permanentAssignsNoCombatDamageLocked(c *Card) bool {
	if c == nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone && scopedEffectHasMod(e, ModAssignsNoCombatDamage) && affectedPredicate(e.Affected)(c, g, nil) {
			return true
		}
	}
	return false
}

// PermanentAssignsNoCombatDamageForEffect is
// permanentAssignsNoCombatDamageLocked by instance ID, for the view's chip
// (ADR 0187 §4). False for a card that is not on the battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) PermanentAssignsNoCombatDamageForEffect(cardID uuid.UUID) bool {
	return g.permanentAssignsNoCombatDamageLocked(findBattlefieldCard(g, cardID))
}

// AssignsNoCombatDamageThisTurnForEffect is "<that creature> assigns no
// combat damage this turn" (ADR 0187 §4, #3569; CR 510.1a). Until cleanup
// the creature assigns 0 combat damage, as attacker and as blocker. Pinned
// to the object (CR 400.7), so a permanent that is flickered comes back
// free of it. Registers nothing for a card that is not on the battlefield
// and returns false for it.
//
// Caller must hold g.mu (write).
func (g *Game) AssignsNoCombatDamageThisTurnForEffect(sourceID, cardID uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil,
		[]Mod{{Kind: ModAssignsNoCombatDamage}}, g.PinnedTo(g.UntilEndOfTurnDuration(), cardID),
		label, timeNowUnixNano())
}
