package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast Mode — {1}{G} Instant:
//
//	"Teamwork 1 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 1 or more.)
//	 Target creature gets +2/+2 and gains trample until end of turn.
//	 Also put a +1/+1 counter on that creature if this spell was cast
//	 using teamwork."
//
// #1703: Teamwork(1). The pump and the trample grant are both
// continuous from announce (CR 611.2c), and the +1/+1 counter — unlike
// them — is permanent, exactly as printed: only the pump and the
// keyword wear off at cleanup. No simplification.
func init() {
	Register(Spec{
		OracleID:      "a465f8a9-d5aa-4dfa-b674-7aa1798cb0bd",
		Name:          "Beast Mode",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(1)},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Beast Mode"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Beast Mode"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.UsedTeamwork() {
				return nil
			}
			return ctx.Game.AddCounterForEffect(t.ID, game.CounterPlusOne, 1)
		},
	})
}
