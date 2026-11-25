package game

import "github.com/google/uuid"

// grant_target_surcharge.go — a cast permission's surcharge that depends
// on the cast spell's targets (#6157): Mavinda, Students' Advocate's "You
// may cast target instant or sorcery card from your graveyard this turn.
// If that spell doesn't target a creature you control, it costs {8} more
// to cast this way."
//
// CR 601.2f: "The player determines the total cost of the spell. Usually
// this is just the mana cost. Some spells have additional or alternative
// costs. Some effects may increase or reduce the cost to pay, or may
// provide other alternative costs." The targets are announced first (CR
// 601.2c), so the increase is read with them: CastPermission.CostMore's
// conditional twin, added by the same pass (grantIncreaseLocked) with the
// announcement's targets in hand, and by the enumerator per target set.
// Data on the permission, not a predicate, so it adds no route from Game
// to a func and survives a restore as a number.

// TargetSurchargeFor is the generic mana `perm`'s
// CostMoreUnlessTargetsYourCreature adds to a cast by `caster` announcing
// `targets`: its N when none of the targets is a creature on the
// battlefield `caster` controls, 0 otherwise (and 0 for every permission
// without one). Caller must hold g.mu.
func (g *Game) TargetSurchargeFor(perm *CastPermission, caster uuid.UUID, targets []TargetRef) int {
	if perm == nil || perm.CostMoreUnlessTargetsYourCreature <= 0 {
		return 0
	}
	for _, t := range targets {
		if t.Kind != TargetCard {
			continue
		}
		if c := findBattlefieldCard(g, t.ID); c != nil && c.Controller == caster && c.IsCreature() {
			return 0
		}
	}
	return perm.CostMoreUnlessTargetsYourCreature
}
