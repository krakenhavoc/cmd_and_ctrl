package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault Skyward — Instant {U} (Scars of Mirrodin, #5577):
//
//	"Target creature gains flying until end of turn. Untap it."
//
// Acrobatic Leap without the pump: the grant, then the untap, both on the
// still-legal target in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2a32a0f-bf06-4cb3-a704-ff0f783f19e4",
		Name:         "Vault Skyward",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Vault Skyward — flying"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
