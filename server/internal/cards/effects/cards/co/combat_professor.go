package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Combat Professor — Creature — Bird Cleric {3}{W}, 2/3 (Jumpstart 2022, #5948):
//
//	"Flying
//	 At the beginning of combat on your turn, target creature you control
//	 gets +1/+0 and gains vigilance until end of turn."
//
// The beginning-of-combat row targets as it goes on the stack and is
// removed when there is no creature to target (CR 603.3d); the pump and the
// keyword are the shared until-end-of-turn effects on the same target. The
// target may be the Professor itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b59665f-23c7-4df9-8e71-f9b1a419063b",
		Name:            "Combat Professor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat("Combat Professor — target creature you control gets +1/+0 and gains vigilance until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Combat Professor — +1/+0"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Combat Professor — vigilance"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "vigilance"))),
		},
	})
}
