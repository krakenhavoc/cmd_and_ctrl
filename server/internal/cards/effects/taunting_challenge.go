package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taunting Challenge — Sorcery, {1}{G}{G}:
//
//	"All creatures able to block target creature this turn do so."
//
// Lure for a turn (#1597): the requirement is a data record pinned to
// the target until end of turn (BlockRequirementUntilEOT), read by the
// engine exactly as Lure's static is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50c52320-9f84-4655-a91a-e09b71f9025c",
		Name:         "Taunting Challenge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementLure,
					Label:  "Taunting Challenge — all creatures able to block it do so",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
