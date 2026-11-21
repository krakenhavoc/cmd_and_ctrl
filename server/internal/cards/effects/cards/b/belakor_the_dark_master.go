package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Be'lakor, the Dark Master — Legendary Creature — Demon Noble {3}{U}{B}{R}, 6/5:
//
//	"Flying
//	 Prince of Chaos — When Be'lakor enters, you draw X cards and you lose X
//	 life, where X is the number of Demons you control.
//	 Lord of Torment — Whenever another Demon you control enters, it deals
//	 damage equal to its power to any target."
//
// X is counted as the ETB resolves, Be'lakor included. The damage is dealt by
// the entering Demon (so its lifelink, deathtouch or protection applies), at
// the power it has as the trigger resolves, last-known if it has gone.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3690ca72-c925-42ab-acc2-7cc1fda05789",
		Name:            "Be'lakor, the Dark Master",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Be'lakor — draw X cards and lose X life (X is the number of Demons you control)", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{Player: ctx.Controller(), N: belakorDemons(ctx)}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Player: ctx.Controller(), Amount: -belakorDemons(ctx)}.Apply(ctx)
				}))),
			{
				Watches: []game.EventKind{game.EventETB},
				Key:     "Be'lakor — the Demon deals damage equal to its power to any target",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature() && c.HasSubtypeWord("Demon")
				},
				Targets: effects.TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					demon, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					ref := ctx.Trigger().Object.Ref()
					return effects.DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: demon.Power}.Apply(ctx)
				},
			},
		},
	})
}

// belakorDemons is X: the Demons the controller controls as the trigger
// resolves.
func belakorDemons(ctx *effects.Context) int {
	return effects.CountYouControl(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.HasSubtypeWord("Demon")
	})
}
