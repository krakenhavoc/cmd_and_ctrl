package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divine Resilience — {W} Instant:
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this
//	 spell.)
//	 Target creature you control gains indestructible until end of
//	 turn. If this spell was kicked, instead any number of target
//	 creatures you control gain indestructible until end of turn."
//
// The kicked clause is the same predicate with an open count (#1716,
// WhenPaid). Each creature still legal at resolution gets its own
// grant; one that left in response is skipped (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "4f6e2e47-34df-4bf3-a546-e06b42840167",
		Name:         "Divine Resilience",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{2}{W}"),
			TargetCreature("any number of target creatures you control", YouControl()).WithCount(0, 0))},
		Targets: TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"indestructible"},
					Label:    "Divine Resilience — indestructible until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
