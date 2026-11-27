package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Negator — Creature — Phyrexian Horror {2}{B}, 5/5 (Urza's Destiny,
// S82 #5621):
//
//	"Trample
//	 Whenever this creature is dealt damage, sacrifice that many permanents."
//
// Barbed Servitor's damage watch: the engine emits one damage event per source, so
// the trigger waits for the damage to finish (AtBatchEnd, ThatMuchDamage) and
// reads the TOTAL the Negator was dealt at the same time, as printed (CR 120.3,
// CR 510.2). As it resolves, the Negator's controller sacrifices that many
// permanents of their choice (Phyrexian Obliterator's N prompts, trimmed after
// every answer); one with fewer permanents sacrifices what they have, and the
// Negator itself may be among them. The trigger still resolves if the damage
// killed the Negator.
//
// No simplification.
func init() {
	const label = "Phyrexian Negator — sacrifice that many permanents"
	effects.Register(effects.Spec{
		OracleID:        "f22fcd17-bc24-41df-81ce-6d11208861b9",
		Name:            "Phyrexian Negator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        label,
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				effects.B17PlayerSacrificesN(g, item.SourceCardID, item.Controller, item.Trigger.Event.Amount, "Phyrexian Negator — sacrifice a permanent")
				return nil
			},
		}},
	})
}
