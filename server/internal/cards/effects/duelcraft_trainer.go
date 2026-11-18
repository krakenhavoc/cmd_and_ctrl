package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duelcraft Trainer — Creature — Human Soldier {3}{W}, 3/3 (Midnight
// Hunt, #5497):
//
//	"First strike
//	 Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, target creature you
//	 control gains double strike until end of turn."
//
// The target is chosen as the trigger goes on the stack, which needs the
// coven to hold then (CR 603.4); with no creature to target the trigger is
// removed (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Duelcraft Trainer — Coven: target creature you control gains double strike until end of turn"
	Register(Spec{
		OracleID:        "81713b4d-85df-4b2c-9673-4f5162590149",
		Name:            "Duelcraft Trainer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(midB05AtBeginningOfCombatIfCoven(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"double strike"}, Label: label}.Apply(ctx)
			}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
