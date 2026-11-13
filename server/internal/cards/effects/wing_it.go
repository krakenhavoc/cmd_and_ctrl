package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wing It — Instant {1}{W} (Modern Horizons 3, #5467):
//
//	"Target creature gets +2/+2 until end of turn. Put a flying counter
//	 on it. Scry 1."
//
// Giant Growth's pump, then Call a Surprise Witness's flying counter
// (CR 122.1b: the creature has flying for as long as the counter stays,
// after the pump has worn off), then the scry, in printed order. A
// target gone by resolution makes the whole spell do nothing, the scry
// included (CR 608.2b). The scry rider is declared on the Purpose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b0b05a0-0be4-498f-be54-8d68b818f90a",
		Name:         "Wing It",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)), Scry: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Wing It — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterFlying, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return Scry{N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
