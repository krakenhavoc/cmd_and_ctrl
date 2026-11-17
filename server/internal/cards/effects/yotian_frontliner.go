package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yotian Frontliner — Artifact Creature — Soldier {1}, 1/1:
//
//	"Whenever this creature attacks, another target creature you control
//	 gets +1/+1 until end of turn.
//	 Unearth {W}"
func init() {
	Register(Spec{
		OracleID:     "cd39e485-71ba-40cb-8608-f5a0dfb7ef5f",
		Name:         "Yotian Frontliner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WheneverThisAttacks("Yotian Frontliner — another target creature you control gets +1/+1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Yotian Frontliner — +1/+1"}.Apply(ctx)
					}),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 1, 1))),
		},
		Activated: []ActivatedAbility{Unearth("{W}")},
	})
}
