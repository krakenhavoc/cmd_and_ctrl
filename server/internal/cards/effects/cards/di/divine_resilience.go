package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4f6e2e47-34df-4bf3-a546-e06b42840167",
		Name:         "Divine Resilience",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{2}{W}"),
			effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0))},
		Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{
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
