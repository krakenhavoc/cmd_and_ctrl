package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Essence — Creature — Wall {1}{W}, 0/4 (Magic 2015, #5556):
//
//	"Defender
//	 Whenever this creature is dealt combat damage, you gain that much life."
//
// Barbed Servitor's damage trigger narrowed to combat damage: the engine emits
// one damage event per source, so the trigger waits for the damage to settle
// (AtBatchEnd, ThatMuchDamage) and gains the TOTAL the Wall was dealt at the
// same time, once (CR 120.3, 510.2, 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1374ac18-01bd-452f-af72-a5f96285ed69",
		Name:            "Wall of Essence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Combat && effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Wall of Essence — you gain that much life",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.S82OldB358GainThatMuchLife,
		}},
	})
}
