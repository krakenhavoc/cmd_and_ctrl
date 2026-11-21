package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunar Frenzy — Instant {X}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"Target creature you control gets +X/+0 and gains first strike and
//	 trample until end of turn."
//
// X is the value announced as the spell is cast (CR 107.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d95fc2cb-5e49-4a22-a93e-d21cf64dcd84",
		Name:         "Lunar Frenzy",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0, "first strike", "trample")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Lunar Frenzy — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "trample"}, Label: "Lunar Frenzy — first strike and trample"}.Apply(ctx)
		},
	})
}
