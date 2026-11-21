package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mouser Attack! — {1}{R} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Choose one —
//	 • Create a 1/1 colorless Robot artifact creature token.
//	 • Target creature gets +3/+0 and gains first strike until end of
//	   turn."
//
// The first bullet makes the set's Robot (Ravenous Robots' row). The second
// is Charge the Sanctum's pump and keyword on the bullet's own target,
// applied only while it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bae5dc68-1fd1-428d-bee7-c238e07b21ea",
		Name:         "Mouser Attack!",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Create a 1/1 colorless Robot artifact creature token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot artifact"), N: 1}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+0 and gains first strike until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Label: "Mouser Attack! — +3/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Mouser Attack! — first strike"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike"))),
		),
	})
}
