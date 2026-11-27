package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b641_helpers.go — shared bodies for Planeshift's non-mana kickers, batch 641 (#5615).

// OldB641KickerSacrificeN is "Kicker—Sacrifice two lands" (Bog Down, Magma Burst):
// KickerSacrifice with the clause's count raised to n, so the caster names exactly
// n permanents and they leave together as the spell is cast (CR 702.33a, 601.2h).
func OldB641KickerSacrificeN(n int, label string, preds ...CardPredicate) game.AdditionalCost {
	c := KickerSacrifice(label, preds...)
	c.Sacrifice = c.Sacrifice.WithCount(n, n)
	return c
}

// OldB641BounceClause returns the pick of target clause `clause` to its owner's
// hand when it is still on the battlefield and legal (CR 608.2b).
func OldB641BounceClause(ctx *Context, clause int) error {
	t, ok := ctx.ClauseTarget(clause)
	if !ok || !OnBattlefield(ctx.Game, t.ID) {
		return nil
	}
	return BounceToHand{Target: t.ID}.Apply(ctx)
}
