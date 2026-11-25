package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gristle Grinner — Creature — Zombie {4}{B}, 3/3 (Jumpstart, #5510):
//
//	"Whenever a creature dies, this creature gets +2/+2 until end of turn."
//
// One trigger per creature that dies, anyone's (each death is its own
// trigger, CR 603.2c). It does nothing for a Grinner that has left the
// battlefield by resolution (a Grinner that is itself the creature that
// died gets nothing).
//
// No simplification.
func init() {
	const label = "Gristle Grinner — this creature gets +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "419f4648-2ace-4bd4-9ae4-8d456688c089",
		Name:         "Gristle Grinner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died
			}, label, func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
