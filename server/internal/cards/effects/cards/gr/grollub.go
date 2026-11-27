package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grollub — Creature — Beast {2}{B}, 3/3 (Exodus, S82 OLD batch 743, #5626):
//
//	"Whenever this creature is dealt damage, each opponent gains that much
//	 life."
//
// Barbed Servitor's damage trigger: the engine emits one damage event per
// source, so it waits for the damage to finish (AtBatchEnd, ThatMuchDamage)
// and each opponent gains the TOTAL the Grollub was dealt at the same time
// (CR 120.3, CR 510.2). Lethal damage still triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22c8c952-2713-418e-8259-aea09b4e740b",
		Name:         "Grollub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Grollub — each opponent gains that much life",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.S82OldB743EachOpponentGainsThatMuch,
		}},
	})
}
