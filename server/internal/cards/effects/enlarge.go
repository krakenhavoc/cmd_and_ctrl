package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enlarge — Sorcery, {3}{G}{G}:
//
//	"Target creature gets +7/+7 and gains trample until end of turn.
//	 It must be blocked this turn if able."
//
// #1684 leftover: BoostUntilEOT, GrantKeywordUntilEOT (trample) and
// the Irresistible Prey shape (BlockRequirementUntilEOT, must-be-
// blocked). No new machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "405cd74a-6b44-4a80-b53c-ff1406d8ff45",
		Name:         "Enlarge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{
					Target:    t.ID,
					Power:     7,
					Toughness: 7,
					Label:     "Enlarge — +7/+7",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"trample"},
					Label:    "Enlarge — trample",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Enlarge — must be blocked this turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
