package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:            "942216f7-6803-4ab3-9239-463002590f6f",
		Name:                "Infused Arrows",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove X charge counters from this artifact: Target creature gets -X/-X until end of turn.",
			Cost:    Plus(TapCost(), RemoveCountersXFromThis(game.CounterCharge, 0)),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
