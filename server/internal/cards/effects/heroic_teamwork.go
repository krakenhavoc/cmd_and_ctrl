package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroic Teamwork — {2}{W} Instant:
//
//	"Teamwork 3 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 3 or more.)
//	 One or two target creatures each get +2/+1 until end of turn. If
//	 this spell was cast using teamwork, draw a card."
//
// #1703: Teamwork(3). The count is one clause with WithCount(1, 2)
// (CR 601.2c), and OnResolve iterates ctx.LegalTargets() so a target
// that left in response is skipped rather than erroring (CR 608.2b) —
// order doesn't matter here, so there's no need to index Targets() by
// slot. No simplification.
func init() {
	Register(Spec{
		OracleID:      "8025313d-048d-4a43-bc01-22398939a7e7",
		Name:          "Heroic Teamwork",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(3)},
		Targets:       TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Heroic Teamwork"}).Apply(ctx); err != nil {
					return err
				}
			}
			if !ctx.UsedTeamwork() {
				return nil
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
