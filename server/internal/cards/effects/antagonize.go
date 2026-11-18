package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Antagonize — Instant {1}{R} (Streets of New Capenna, #5491):
//
//	"Target creature gets +4/+3 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4e860f1-e637-4640-b257-57c2e42793ce",
		Name:         "Antagonize",
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 3)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: 3, Label: "Antagonize — +4/+3"}.Apply(ctx)
		},
	})
}
