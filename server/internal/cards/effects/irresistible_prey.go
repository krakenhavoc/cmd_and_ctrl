package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irresistible Prey — Sorcery, {G}:
//
//	"Target creature must be blocked this turn if able.
//	 Draw a card."
//
// The requirement is a data record pinned to the target until end of
// turn (#1597, BlockRequirementUntilEOT), so it survives undo and a
// restore point and does not follow the creature if it leaves and
// returns (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb102767-f8d8-4477-9a24-07686870bda7",
		Name:         "Irresistible Prey",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Irresistible Prey — must be blocked this turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
