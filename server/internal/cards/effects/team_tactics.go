package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Team Tactics — {1}{R} Instant:
//
//	"Teamwork 1 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 1 or more.)
//	 Target creature gains double strike until end of turn. If this
//	 spell was cast using teamwork, that creature also gains trample
//	 until end of turn."
//
// #1703: Teamwork(1). Both keywords are the same Layer 6 grant, so a
// teamwork cast appends "trample" to the same keyword set rather than
// stacking a second effect. No simplification.
func init() {
	Register(Spec{
		OracleID:      "98b99d7c-d006-471a-8583-0468743c1597",
		Name:          "Team Tactics",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(1)},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			kws := []string{"double strike"}
			if ctx.UsedTeamwork() {
				kws = append(kws, "trample")
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: kws, Label: "Team Tactics"}.Apply(ctx)
		},
	})
}
