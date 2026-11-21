package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Laid to Rest — Enchantment {3}{G} (Crimson Vow):
//
//	"Whenever a Human you control dies, draw a card.
//	 Whenever a creature you control with a +1/+1 counter on it dies, you gain
//	 2 life."
//
// Both are looks-back triggers (CR 603.10a): "a Human" is judged on the
// creature's last-known subtypes and controller, "with a +1/+1 counter on it"
// on its last-known counters. A Human that dies with a counter triggers both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae6dc45c-b8c9-4cbf-892f-8c00f4199aaf",
		Name:         "Laid to Rest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Human"))
			}, "Laid to Rest — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Laid to Rest — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
