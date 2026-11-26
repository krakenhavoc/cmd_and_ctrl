package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horde Ambusher — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Human Berserker {1}{R}, 2/2: Whenever this creature blocks, it deals 1
//	damage to you. Morph—Reveal a red card in your hand. When this creature is
//	turned face up, target creature can't block this turn.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "fb7f45eb-c134-4493-af05-e3b14d80e48c",
		Name:             "Horde Ambusher",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MorphReveal("R")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			}, "Horde Ambusher — it deals 1 damage to you", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				return effects.DealDamage{SourceObject: &ref, Target: item.Controller, Amount: 1}.Apply(ctx)
			}),
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Horde Ambusher — target creature can't block this turn",
					effects.TargetCreatureCantBlockThisTurn("Horde Ambusher — can't block")),
				effects.TargetCreature("target creature")),
		},
	})
}
