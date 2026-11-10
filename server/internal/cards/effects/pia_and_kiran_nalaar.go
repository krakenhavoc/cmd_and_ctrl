package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pia and Kiran Nalaar — Legendary Creature — Human Artificer {2}{R}{R},
// 2/2:
//
//	"When Pia and Kiran Nalaar enters, create two 1/1 colorless Thopter
//	 artifact creature tokens with flying.
//	 {2}{R}, Sacrifice an artifact: Pia and Kiran Nalaar deals 2 damage to
//	 any target."
//
// Whirler Rogue's two Thopters and Goblin Bombardment's sacrifice-for-damage
// shape: the artifact is a cost paid on announce (it may be a Thopter, or
// the one the damage would otherwise be waiting on), and the damage comes
// from the creature, last-known information if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86084ed9-b8bb-4289-9579-6056194787bc",
		Name:         "Pia and Kiran Nalaar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pia and Kiran Nalaar — create two 1/1 colorless Thopter artifact creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice an artifact: Pia and Kiran Nalaar deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{2}{R}"), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := NewContext(g, item)
				return DealDamage{
					Source: ctx.Source(),
					Target: item.Targets[0].ID,
					Amount: 2,
				}.Apply(ctx)
			},
		}},
	})
}
