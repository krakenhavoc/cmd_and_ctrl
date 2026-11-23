package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yotian Frontliner — Artifact Creature — Soldier {1}, 1/1:
//
//	"Whenever this creature attacks, another target creature you control
//	 gets +1/+1 until end of turn.
//	 Unearth {W}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd39e485-71ba-40cb-8608-f5a0dfb7ef5f",
		Name:         "Yotian Frontliner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WheneverThisAttacks("Yotian Frontliner — another target creature you control gets +1/+1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Yotian Frontliner — +1/+1"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{W}")},
	})
}
