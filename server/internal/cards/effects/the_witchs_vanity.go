package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Witch's Vanity — Enchantment — Saga {1}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy target creature an opponent controls with mana value 2
//	 or less.
//	 II — Create a Food token.
//	 III — Create a Wicked Role token attached to target creature you
//	 control."
//
// Chapters I and III target, so the engine computes the legal set as
// each goes on the stack and drops one with none (CR 603.3d). Chapter
// III is not "up to one": with no creature of yours it does nothing.
//
// No simplifications.
func init() {
	const name = "The Witch's Vanity"
	Register(Spec{
		OracleID:     "ad2793f3-a36f-4a09-a38f-eb98ce0a2b79",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "destroy target creature an opponent controls with mana value 2 or less"),
				TargetCreature("target creature an opponent controls with mana value 2 or less", OpponentControls(), ManaValueLE(2)),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if id, ok := b16FirstLegalTargetCard(ctx); ok {
						return DestroyTarget{Target: id}.Apply(ctx)
					}
					return nil
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a Food token"),
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "create a Wicked Role token attached to target creature you control"),
				TargetCreature("target creature you control", YouControl()),
				createRoleOnFirstTarget(RoleWicked)),
		},
	})
}
