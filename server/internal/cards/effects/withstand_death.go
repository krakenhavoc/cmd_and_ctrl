package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Withstand Death — Instant {G} (Scars of Mirrodin, #5577):
//
//	"Target creature gains indestructible until end of turn."
//
// Ephemeral Shields' grant without the convoke.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c462adf3-4ade-4b90-bdec-7247b080471f",
		Name:         "Withstand Death",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Withstand Death — indestructible"}.Apply(ctx)
		},
	})
}
