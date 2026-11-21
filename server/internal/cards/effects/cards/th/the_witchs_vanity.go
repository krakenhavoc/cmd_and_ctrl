package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ad2793f3-a36f-4a09-a38f-eb98ce0a2b79",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "destroy target creature an opponent controls with mana value 2 or less"),
				effects.TargetCreature("target creature an opponent controls with mana value 2 or less", effects.OpponentControls(), effects.ManaValueLE(2)),
				effects.DestroyTheFirstLegalTarget),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a Food token"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.FoodToken(), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "create a Wicked Role token attached to target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.CreateRoleOnFirstTarget(effects.RoleWicked)),
		},
	})
}
