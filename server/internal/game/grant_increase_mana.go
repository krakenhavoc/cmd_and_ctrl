package game

import "github.com/google/uuid"

// grant_increase_mana.go — #6091: a cast permission's coloured additional
// cost, "you may cast target instant or sorcery card from your graveyard
// by paying {R}{R} in addition to its other costs" (Ogre Battlecaster).
//
// CastPermission.CostMore adds generic mana only. CostMoreMana names the
// mana, coloured or not, and is added to the total cost as an additional
// cost (CR 601.2f) by the same pass that adds CostMore, so the payment,
// the preview and the bot's enumerator agree.

// CostMoreManaOf is the parsed CostMoreMana of perm, or nil when it names
// none or does not parse (a malformed permission grants less, never more:
// the cast it would price is refused by the cast path's own parse of it,
// see grantIncreaseManaLocked).
func CostMoreManaOf(perm *CastPermission) *ParsedCost {
	if perm == nil || perm.CostMoreMana == "" {
		return nil
	}
	pc, err := ParseCost(perm.CostMoreMana)
	if err != nil {
		return nil
	}
	return &pc
}

// grantIncreaseManaLocked is CostQuery.GrantIncreaseMana for one
// announcement: the CostMoreMana of the permission the cast claims, or nil
// for a cast that uses none. grantIncreaseLocked's twin, asking the same
// claim. A hand or command-zone cast claims no stored permission. Caller
// must hold g.mu.
func (g *Game) grantIncreaseManaLocked(playerID uuid.UUID, card Card, from ZoneKind, params CastSpellParams) *ParsedCost {
	if from == ZoneHand || from == ZoneCommand {
		return nil
	}
	return CostMoreManaOf(g.permissionForSpellLocked(g.CastPermissionForFaceLocked(playerID, card, from, params.AlternativeCost, params.Face), card))
}

// increaseByGrantMana adds a permission's additional mana (nil: none).
func increaseByGrantMana(cost ParsedCost, extra *ParsedCost) ParsedCost {
	if extra == nil {
		return cost
	}
	return increaseBy(cost, 1, extra)
}
