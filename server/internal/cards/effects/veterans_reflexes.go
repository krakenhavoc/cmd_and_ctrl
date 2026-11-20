package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran's Reflexes — Instant {W} (Worldwake, #5582):
//
//	"Target creature gets +1/+1 until end of turn. Untap that creature."
//
// The pump, then the untap, on the target still legal as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efee0ca0-0897-4972-99c4-432561b6e9f8",
		Name:         "Veteran's Reflexes",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Veteran's Reflexes — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
