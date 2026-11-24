package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kira, Great Glass-Spinner — Legendary Creature — Spirit {1}{U}{U}, 2/2
// (Secret Lair Drop, #5462):
//
//	"Flying
//	 Creatures you control have "Whenever this creature becomes the target
//	 of a spell or ability for the first time each turn, counter that spell
//	 or ability.""
//
// Nadu, Winged Wisdom's grant: the trigger belongs to each creature (ADR 0093)
// and its own label is counted per object, so each creature counters the first
// spell or ability aimed at it each turn, of any controller, and the second
// one is not countered. The event carries the stack object, and CounterTarget
// counters a spell or an ability alike (CR 701.6). A countered spell or ability
// that can't be countered stays. The trigger counts from the moment it is put
// on the stack, so one that was countered or responded to still used the turn's
// first targeting.
//
// No simplification.
func init() {
	const (
		key   = "kira-great-glass-spinner/counter-first-targeting"
		label = "Kira, Great Glass-Spinner — counter that spell or ability"
		text  = "Whenever this creature becomes the target of a spell or ability for the first time each turn, counter that spell or ability."
	)
	effects.Register(effects.Spec{
		OracleID:        "60808c8f-5257-451b-a9b6-389ba51598d5",
		Name:            "Kira, Great Glass-Spinner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.Self(ev, source, lki, g) && g.TriggeredThisTurn(source.InstanceID, label) < 1
				}, label, func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.CounterTarget{StackID: item.Trigger.Event.StackItemID}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, key), "Creatures you control have \""+text+"\"")},
	})
}
