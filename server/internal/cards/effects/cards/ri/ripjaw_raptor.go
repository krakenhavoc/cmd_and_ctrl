package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ripjaw Raptor — Creature — Dinosaur {2}{G}{G}, 4/5:
//
//	"Enrage — Whenever this creature is dealt damage, draw a card."
//
// Strong, the Brutish Thespian's enrage shape: one trigger per damage
// instance, not per source (CR 120.3, CR 510.2), waiting for the damage
// to settle (AtBatchEnd). Lethal damage still triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80bb9ce4-1f3c-42ca-b75f-9a6b2f42073b",
		Name:         "Ripjaw Raptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Ripjaw Raptor — Enrage: draw a card",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
