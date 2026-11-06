package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stand United — Instant {1}{G/W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Target creature gets +2/+2 until end of turn. If it's an Ally,
//	 scry 2."
//
// "If it's an Ally" reads the creature as the spell resolves, through the
// layers (an Ally granted by a type-changing effect counts), and "Ally"
// is a word of the spell's text (CR 612.2). A target that has become
// illegal makes the whole spell do nothing (CR 608.2b), the scry
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a05ee6e-1812-4792-ad00-6df8fa4fd85f",
		Name:         "Stand United",
		Completeness: CompletenessFull,
		Purpose:      WithRiders(ForTargets(BenefitUntilEOT(0, 2, 2)), game.Purpose{Scry: 2}),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: 2, Toughness: 2, Label: "Stand United — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok || !c.HasSubtype(ctx.TextSubtype("Ally")) {
				return nil
			}
			return Scry{N: 2}.Apply(ctx)
		},
	})
}
