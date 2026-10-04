package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// doubling.go — "double <a creature>'s power [and toughness]" (CR
// 701.10), shared by Unleash Fury, The Skullspore Nexus and Zopandrel,
// Hunger Dominus (S58 PR 6).
//
// CR 701.10a: doubling is a continuous effect that MODIFIES power and
// toughness (layer 7c), never one that sets them. CR 701.10b: the
// creature gets +X/+0, where X is its power as the spell or ability
// resolves, so it is a one-time reading and a pump later in the turn
// is not doubled again. CR 701.10c: a negative power doubles downward
// (-X/-0), which is the same arithmetic as adding the unclamped power,
// so X is PowerForComparison (layers and counters, no clamp) and the
// toughness half is CurrentToughness, which is not clamped either.
//
// Append-only.

// doubledBy is the +X/+Y one creature gets from doubling, read now.
// `toughness` false is "double its power" and leaves Y at zero. The
// second result is false when the object is not a creature on the
// battlefield any more, in which case nothing is doubled.
func doubledBy(g *game.Game, id uuid.UUID, toughness bool) (int, int, bool) {
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(id)
	if !ok || !onBattlefield(g, id) || !c.IsCreature() {
		return 0, 0, false
	}
	dt := 0
	if toughness {
		dt = c.CurrentToughness()
	}
	return c.PowerForComparison(), dt, true
}

// DoublePowerUntilEOT is "double target creature's power until end of
// turn" (Unleash Fury, Bulk Up's wording; The Skullspore Nexus's
// activated ability).
func DoublePowerUntilEOT(ctx *Context, id uuid.UUID, label string) error {
	dp, _, ok := doubledBy(ctx.Game, id, false)
	if !ok {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: dp, Label: label}.Apply(ctx)
}

// DoublePowerAndToughnessOfEachUntilEOT is "double the power and
// toughness of each <creature> until end of turn" (Zopandrel). Every
// creature's X and Y are read before any of them is changed, so the
// order the set is walked in cannot matter, and each gets its own
// +X/+Y (CR 701.10b). The set is locked now (CR 611.2c).
func DoublePowerAndToughnessOfEachUntilEOT(ctx *Context, ids []uuid.UUID, label string) error {
	type bump struct {
		id     uuid.UUID
		dp, dt int
	}
	bumps := make([]bump, 0, len(ids))
	for _, id := range ids {
		dp, dt, ok := doubledBy(ctx.Game, id, true)
		if ok {
			bumps = append(bumps, bump{id: id, dp: dp, dt: dt})
		}
	}
	for _, b := range bumps {
		if err := (BoostUntilEOT{Target: b.id, Power: b.dp, Toughness: b.dt, Label: label}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
