package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9d835137-584c-46bd-952e-4b2812bba1a3",
		Name:         "Charge the Sanctum",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Creatures you control get +2/+0 until end of turn."),
			Mode("Target creature gets +2/+0 and gains first strike until end of turn. Put a +1/+1 counter on it.",
				TargetCreature("target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 2,
					Label: "Charge the Sanctum — +2/+0",
				}.Apply(ctx)
			case ctx.HasMode(1):
				t, ok := ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Charge the Sanctum — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"first strike"},
					Label:    "Charge the Sanctum — first strike",
				}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
