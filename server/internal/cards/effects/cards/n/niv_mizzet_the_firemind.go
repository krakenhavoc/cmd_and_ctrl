package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niv-Mizzet, the Firemind — Legendary Creature — Dragon Wizard
// {2}{U}{U}{R}{R}, 4/4 (EDHREC rank 2524):
//
//	"Flying
//	 Whenever you draw a card, Niv-Mizzet deals 1 damage to any
//	 target.
//	 {T}: Draw a card."
//
// The original Firemind: every card you draw is a ping, and he draws
// one himself. Flying rides PrintedKeywords. The draw trigger is
// Sheoldred's own-draw condition with a target — EventDrawCard fires
// once per card, so a seven-card draw is seven triggers, each with
// its own "any target" pick (a player, a creature, a planeswalker, a
// battle) chosen as it goes on the stack. The tap is a CR 602
// activation with a tap component, so summoning sickness applies
// and the drawn card fires the trigger, as printed — Curiosity's
// loop is a real loop here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "959acb66-84ca-4535-bca2-ad591895735e",
		Name:            "Niv-Mizzet, the Firemind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Key:     "Niv-Mizzet, the Firemind — deal 1 damage to any target",
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
