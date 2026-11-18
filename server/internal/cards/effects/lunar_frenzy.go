package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunar Frenzy — Instant {X}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"Target creature you control gets +X/+0 and gains first strike and
//	 trample until end of turn."
//
// X is the value announced as the spell is cast (CR 107.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d95fc2cb-5e49-4a22-a93e-d21cf64dcd84",
		Name:         "Lunar Frenzy",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(PowerXUntilEOT(0, "first strike", "trample")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Lunar Frenzy — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "trample"}, Label: "Lunar Frenzy — first strike and trample"}.Apply(ctx)
		},
	})
}
