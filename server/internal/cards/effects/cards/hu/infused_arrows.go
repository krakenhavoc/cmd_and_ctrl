package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infused Arrows — Artifact {4}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 {T}, Remove X charge counters from this artifact: Target creature
//	 gets -X/-X until end of turn."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). X is the number of counters removed as the cost is
// paid, which the effect reads back.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "942216f7-6803-4ab3-9239-463002590f6f",
		Name:                "Infused Arrows",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove X charge counters from this artifact: Target creature gets -X/-X until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterCharge, 0)),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
