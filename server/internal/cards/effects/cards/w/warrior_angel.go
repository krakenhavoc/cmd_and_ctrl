package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warrior Angel — Creature — Angel Warrior {4}{W}{W}, 3/4 (Stronghold, S82 OLD
// batch 747, #5627):
//
//	"Flying
//	 Whenever this creature deals damage, you gain that much life."
//
// Exalted Angel's reading of "deals damage": each damage event the Angel is
// the source of, with the amount actually dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "97140551-341a-40a1-93fe-0fcaa0c3be33",
		Name:            "Warrior Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Warrior Angel — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
