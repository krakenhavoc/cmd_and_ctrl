package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// variable_sacrifice_cost.go — ADR 0100 §3 (#1732, sub-PR 4): the
// shared bodies of the cards whose additional cost sacrifices a
// VARIABLE number of permanents. The constructors themselves
// (SacrificeXCost, SacrificeAnyNumberCost) sit with the other
// additional costs in additional_cost.go, and the per-sacrifice discount
// (CostsLessPerSacrificed) with the self cost modifiers. Append-only.
//
// Every reader here reads the ANNOUNCEMENT, not the board: by
// resolution the sacrificed permanents are in graveyards (or, for a
// token, nowhere), so "for each creature sacrificed this way" is
// PaidCost.Sacrificed through ctx.Sacrificed(), and "sacrifice X" is the
// announced X through ctx.X().

// vsDestroyTheTargets destroys every target that is still legal as ONE
// simultaneous event (CR 608.2b skips the rest) — Eliminate the
// Competition's "destroy X target creatures", Immoral Bargain's
// "destroy X target nonland permanents". One event, so a Blood Artist
// sees every death at once rather than one per target.
func vsDestroyTheTargets(ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) > 0 {
		ctx.Game.DestroyPermanentsForEffect(ids)
	}
	return nil
}

// vsPumpTheTarget gives the spell's one creature target +power/+toughness
// until end of turn, when it is still legal and the boost is not zero —
// Vicious Betrayal's "+2/+2 for each creature sacrificed this way",
// Devouring Rage's "+3/+0, and +3/+0 more for each Spirit".
func vsPumpTheTarget(ctx *Context, power, toughness int, label string) error {
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	return BoostUntilEOT{Target: legal[0].ID, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
}

// vsXTarget is "X target <permanents>" — the target clause whose count
// is the announced X, which on these cards is the sacrifice count
// (CR 107.3i: every X on the object is one number).
func vsXTarget(label string, preds ...CardPredicate) *game.TargetSpec {
	spec := TargetPermanent(label, preds...)
	spec.CountFromX = true
	return spec
}
