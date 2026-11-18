package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blessed Defiance — Instant {W} (Innistrad: Midnight Hunt, #5497):
//
//	"Target creature you control gets +2/+0 and gains lifelink until end of
//	 turn. When that creature dies this turn, create a 1/1 white Spirit
//	 creature token with flying."
//
// Felonious Rage's shape: the pump and the grant, then a delayed triggered
// ability bound to that one object (CR 603.7a), controlled by the spell's
// controller (CR 603.7d), so the Spirit is theirs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04e3d36f-5dec-422c-a371-15e135fdface",
		Name:         "Blessed Defiance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "lifelink")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Blessed Defiance — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: "Blessed Defiance — lifelink"}).Apply(ctx); err != nil {
				return err
			}
			return whenThatCreatureDiesThisTurn(ctx, t.ID, "Blessed Defiance — create a 1/1 Spirit with flying", s82MidSpiritBody)
		},
	})
}
