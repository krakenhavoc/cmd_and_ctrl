package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruby Collector — Creature — Human Detective {W}, 1/1 (Mystery Booster
// Commander, #5948):
//
//	"When you attack with three or more creatures, conjure a card named
//	 Mox Ruby into your hand. This ability triggers only once.
//	 {1}{R}: Creatures you control get +1/+0 until end of turn."
//
// "Attack with three or more creatures" is Phoenix Chick's batch count
// over your attack declaration. The conjured Mox Ruby is a real card
// (ADR 0099 amendment 2026-11-24).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca927d6b-77f7-49fa-a686-12fa981a8bd4",
		Name:         "Ruby Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnce(effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Ruby Collector — conjure a card named Mox Ruby into your hand (you attacked with three or more creatures)",
				effects.Do(effects.CheckConjure("Ruby Collector", effects.Conjure{Names: []string{"Mox Ruby"}, Zone: game.ZoneHand}))),
				effects.ThatManyEventsAtLeast(3))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
					Label: "Ruby Collector — creatures you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
