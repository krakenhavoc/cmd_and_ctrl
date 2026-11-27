package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnfluke — Creature — Elemental {3}{W}, 0/3 (Lorwyn, S82 OLD, #5593):
//
//	"Flash
//	 When this creature enters, prevent the next 3 damage that would be dealt to
//	 any target this turn.
//	 Evoke {W}"
//
// Withstand's three-point prevention shield on the target (a creature, player
// or planeswalker), re-checked as the trigger resolves, under flash and
// Mulldrifter's evoke.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9ca9168f-54cb-4091-b80c-7c264773c1a2",
		Name:             "Dawnfluke",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dawnfluke — prevent the next 3 damage that would be dealt to any target this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if len(ctx.LegalTargets()) == 0 || len(item.Targets) == 0 {
							return nil
						}
						return effects.PreventNextDamage{Target: item.Targets[0].ID, Amount: 3, Label: "Dawnfluke: prevent the next 3 damage"}.Apply(ctx)
					}),
				effects.TargetAny()),
		},
	})
}
