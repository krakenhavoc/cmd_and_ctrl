package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charge the Sanctum — Instant {2}{R/W}:
//
//	"Choose one —
//	 • Creatures you control get +2/+0 until end of turn.
//	 • Target creature gets +2/+0 and gains first strike until end of
//	   turn. Put a +1/+1 counter on it."
//
// Mode one fixes its set of creatures as the spell resolves
// (CR 611.2c). Mode two applies the pump, the keyword and the counter
// to the one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d835137-584c-46bd-952e-4b2812bba1a3",
		Name:         "Charge the Sanctum",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Creatures you control get +2/+0 until end of turn."),
			effects.ModeWithPurpose(effects.Mode("Target creature gets +2/+0 and gains first strike until end of turn. Put a +1/+1 counter on it.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike"))),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 2,
					Label: "Charge the Sanctum — +2/+0",
				}.Apply(ctx)
			case ctx.HasMode(1):
				t, ok := effects.ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Charge the Sanctum — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"first strike"},
					Label:    "Charge the Sanctum — first strike",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
