package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const haradrimSpearmasterLabel = "Haradrim Spearmaster — another target creature you control gets +1/+0 until end of turn"

// Haradrim Spearmaster — Creature — Human Warrior {2}{R}, 2/3 (The Lord of
// the Rings: Tales of Middle-earth Commander):
//
//	"Reach
//	 At the beginning of combat on your turn, another target creature you
//	 control gets +1/+0 until end of turn."
//
// Cantankerous Captain's targeted combat trigger at +1/+0, with "another"
// keeping the Spearmaster off its own clause. Reach is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4da49285-e577-487f-9997-ba49f2b5753e",
		Name:            "Haradrim Spearmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat(haradrimSpearmasterLabel, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 1, Label: "Haradrim Spearmaster — +1/+0"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0))),
		},
	})
}
