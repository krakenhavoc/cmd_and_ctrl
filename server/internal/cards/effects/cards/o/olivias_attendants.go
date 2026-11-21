package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Olivia's Attendants — Creature — Vampire {4}{R}{R}, 6/6 (Innistrad:
// Crimson Vow, #5495):
//
//	"Menace
//	 Whenever this creature deals damage, create that many Blood tokens.
//	 {2}{R}: This creature deals 1 damage to any target."
//
// El-Hajjâj's reading of "deals damage": every damage event the creature is
// the source of, combat or not, with the amount actually dealt, so its own
// {2}{R} ping makes one Blood token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b384c3c-0d6b-4bba-b4c7-5af8b32bdf1c",
		Name:            "Olivia's Attendants",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Olivia's Attendants — create that many Blood tokens",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.BloodToken(), N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
