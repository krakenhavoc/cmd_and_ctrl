package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Pits of Rath — Enchantment {3}{B}{B} (Eighth Edition, #5607):
//
//	"Whenever a creature is dealt damage, destroy it. It can't be
//	 regenerated."
//
// One trigger per damage event (CR 120.3), whichever creature is hurt and
// whoever controls it; it names the creature as it was (CR 400.7), so a
// creature that already died of the damage is not destroyed twice.
//
// No simplification.
func init() {
	const label = "Death Pits of Rath — destroy that creature, it can't be regenerated"
	effects.Register(effects.Spec{
		OracleID:     "aead53ef-ffcf-4ae2-ab81-40aa9627adc9",
		Name:         "Death Pits of Rath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B28CreatureWasDealtDamage(ev, g)
				return ok
			},
			Key:    label,
			Build:  effects.BuildNamingTheEventTarget(label),
			Effect: effects.S82OldB615DestroyTheNamedCreatureNoRegen,
		}},
	})
}
