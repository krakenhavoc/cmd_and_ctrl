package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodline Culling — Instant {1}{B}{B} (Innistrad: Midnight Hunt, #5497):
//
//	"Choose one —
//	 • Target creature gets -5/-5 until end of turn.
//	 • Creature tokens get -2/-2 until end of turn."
//
// Abrade's modal shape. The second mode is Cry of the Carnarium's
// -2/-2 over a set read once as it resolves (CR 611.2c), narrowed to
// tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b7fabe8-bc44-47b1-93bb-b895a1bd0aa9",
		Name:         "Bloodline Culling",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Target creature gets -5/-5 until end of turn.", TargetCreature("target creature")),
				ForTargets(HarmUntilEOT(0, -5, -5))),
			ModeWithPurpose(Mode("Creature tokens get -2/-2 until end of turn."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}}),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: item.Targets[0].ID, Power: -5, Toughness: -5, Label: "Bloodline Culling — -5/-5 until end of turn"}.Apply(ctx)
			case ctx.HasMode(1):
				return BoostUntilEOT{Match: And(Creature(), IsTokenPredicate()), Power: -2, Toughness: -2, Label: "Bloodline Culling — creature tokens get -2/-2 until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
