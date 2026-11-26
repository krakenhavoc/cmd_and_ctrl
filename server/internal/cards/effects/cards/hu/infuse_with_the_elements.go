package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infuse with the Elements — Instant {3}{G} (Battle for Zendikar):
//
//	"Converge — Put X +1/+1 counters on target creature, where X is the
//	 number of colors of mana spent to cast this spell. That creature gains
//	 trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "beea5db2-1771-417f-9a31-028b3d31aa66",
		Name:                "Infuse with the Elements",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Targets:             effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: ctx.ColorsSpentCount()}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Infuse with the Elements — trample"}.Apply(ctx)
			}
			return nil
		},
	})
}
