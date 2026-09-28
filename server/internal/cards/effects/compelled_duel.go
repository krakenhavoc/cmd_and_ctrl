package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compelled Duel — Sorcery, {1}{G}:
//
//	"Target creature gets +3/+3 until end of turn and must be blocked
//	 this turn if able."
//
// #1684 leftover: BoostUntilEOT plus the Irresistible Prey shape
// (BlockRequirementUntilEOT, must-be-blocked). No new machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "256cc907-2a13-4f43-9d37-2e52034cdaf4",
		Name:         "Compelled Duel",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{
					Target:    t.ID,
					Power:     3,
					Toughness: 3,
					Label:     "Compelled Duel — +3/+3",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Compelled Duel — must be blocked this turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
