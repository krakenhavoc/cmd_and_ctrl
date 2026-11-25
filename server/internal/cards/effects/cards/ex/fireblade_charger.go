package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fireblade Charger — Creature — Goblin Warrior {R}, 1/1 (Zendikar Rising,
// S82 OLD #5508):
//
//	"As long as this creature is equipped, it has haste.
//	 When this creature dies, it deals damage equal to its power to any
//	 target."
//
// The damage is its power as it last existed (CR 608.2h), dealt by the
// Charger as it was on the battlefield.
//
// No simplification.
func init() {
	haste := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && effects.IsEquippedCreature(g, *target)
	}, "haste")
	haste.Label = "As long as this creature is equipped, it has haste."
	effects.Register(effects.Spec{
		OracleID:     "4d83ad77-dbac-4a03-8034-43d950d0a18c",
		Name:         "Fireblade Charger",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{haste},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Fireblade Charger — it deals damage equal to its power to any target",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						info, _ := ctx.TriggeringPermanent()
						if info.Power <= 0 {
							return nil
						}
						for _, t := range ctx.LegalTargets() {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: info.Power}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetAny()),
		},
	})
}
