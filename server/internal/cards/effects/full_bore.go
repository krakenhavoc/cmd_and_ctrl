package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Full Bore — Instant {R} (Edge of Eternities, #4461):
//
//	"Target creature you control gets +3/+2 until end of turn. If that
//	 creature was cast for its warp cost, it also gains trample and haste
//	 until end of turn."
//
// Yip Yip!'s pump-then-check shape. "Cast for its warp cost" is the
// permanent's cast provenance (CR 400.7d): the alternative cost key the
// spell that became it was paid with, here warp's (CR 702.185). A
// creature recast from exile after its warp ended was cast for its mana
// cost, and a token or reanimated creature was not cast at all, so
// neither gains the keywords. A target no longer legal gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b96d003-d3b4-4181-9f50-06a3e6eb1916",
		Name:         "Full Bore",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: 3, Toughness: 2, Label: "Full Bore — +3/+2"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok || c.Provenance.AltCost != Warp("").Key {
				return nil
			}
			return GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample", "haste"}, Label: "Full Bore — trample and haste"}.Apply(ctx)
		},
	})
}
