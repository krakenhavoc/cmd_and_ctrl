package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiteful Hexmage — Creature — Human Warlock {1}{B}:
//
//	"When this creature enters, create a Cursed Role token attached to
//	 target creature you control. (If you control another Role on it,
//	 put that one into the graveyard. Enchanted creature is 1/1.)"
//
// The target may be the Hexmage itself. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d07e0d88-8b7f-423d-97b7-cf7b80d8924c",
		Name:         "Spiteful Hexmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Spiteful Hexmage — create a Cursed Role token attached to target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return CreateRoleToken{Role: RoleCursed, Host: id}.Apply(ctx)
				}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
