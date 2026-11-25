package game

import "github.com/google/uuid"

// sacrifice_cost_bar.go — "Players can't pay life or sacrifice nonland
// permanents to cast spells or activate abilities." (Yasharn,
// Implacable Earth, #6167).
//
// A world rule its source carries on itself as the second-set bit
// PlayersCantPayLifeOrSacrificeNonlandToCastOrActivate (restrictions.go),
// derived from the battlefield on every question, so two sources compose
// and one leaving revokes nothing.
//
// CR 101.2: "can't" beats "can", and CR 118.3: a cost that can't be paid
// can't be paid. So while the rule is live a nonland permanent can pay
// no sacrifice COST: a spell's additional cost (casualty, bargain, "As an
// additional cost, sacrifice a creature"), an alternative cost (emerge),
// an activated ability's "Sacrifice this" or "Sacrifice a creature", and
// a mana ability's (a Treasure, Ashnod's Altar), CR 601.2h and CR 602.2b
// alike. A land still may: Yasharn names nonland permanents only, so a
// fetch land works.
//
// Every cost walk already asks one question of the permanent it would
// sacrifice, "can this be sacrificed" (#3275, cant_be_sacrificed.go).
// sacrificeCostBarredLocked is that question widened by this rule, and
// replaces it at the COST sites only: validateSacrificeCostLocked (the
// one check activations, mana abilities and additional costs share), the
// candidate list the view, the enumerator and the payability checks
// read, the alternative-cost walk, the sacrifice-the-grantor and
// sacrifice-all costs, and the auto-tapper. An effect's "sacrifice a
// creature" (an edict) and the sacrifice route itself still ask the
// narrower question: they are not costs.
//
// The life half is life_cost_bar.go's: cantPayLifeToCastOrActivateLocked
// reads this bit too, and CanPayLifeForManaAbilityLocked adds the mana
// abilities Karn's Sylex leaves alone.

// cantPayLifeOrSacrificeNonlandLocked reports whether a permanent on the
// battlefield carries the rule, read off the layered view. A phased-out
// source is not in the battlefield slice and stops nothing (CR 702.26b).
//
// Caller must hold g.mu (read or write).
func (g *Game) cantPayLifeOrSacrificeNonlandLocked() bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if ExtraRestrictionsOn(&g.Battlefield.Cards[i]).Has(PlayersCantPayLifeOrSacrificeNonlandToCastOrActivate) {
			return true
		}
	}
	return false
}

// sacrificeCostBarredLocked reports whether the permanent `c` can't be
// sacrificed to pay a cost right now: it can't be sacrificed at all
// (#3275), or it is a nonland permanent while Yasharn's rule is live.
// False for nil.
//
// Caller must hold g.mu.
func (g *Game) sacrificeCostBarredLocked(c *Card) bool {
	if c == nil {
		return false
	}
	if g.permanentCantBeSacrificedLocked(c) {
		return true
	}
	return !c.IsLand() && g.cantPayLifeOrSacrificeNonlandLocked()
}

// SacrificeCostBarredForEffect is sacrificeCostBarredLocked by instance
// ID, for the legal-move enumerator and the view. False for a card that
// is not on the battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) SacrificeCostBarredForEffect(cardID uuid.UUID) bool {
	return g.sacrificeCostBarredLocked(findBattlefieldCard(g, cardID))
}
