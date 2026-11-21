package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Bond — Enchantment {2}{G} (EDHREC rank 559):
//
//	"Whenever a creature you control with power 3 or greater enters,
//	 draw a card."
//
// Garruk's Uprising's ongoing half on its own: a card for every big
// creature. The power is read as the creature enters (CurrentPower:
// effective power plus counters, so a creature that arrives with
// +1/+1 counters counts them).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd",
		Name:         "Elemental Bond",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() >= 3
			}, "Elemental Bond — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
