package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Lunge — Instant {1}{W} (Foundations Jumpstart, #5929):
//
//	"Target creature gets +2/+2 and gains flying until end of turn. You gain
//	 2 life."
//
// The boost and the keyword, then the life. A target that is gone as the
// spell resolves fizzles the whole spell, life included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87644a09-241a-44e5-99f2-16fc9a7d8b93",
		Name:         "Desperate Lunge",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Toughness: 2, Label: "Desperate Lunge — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: "Desperate Lunge — flying"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
