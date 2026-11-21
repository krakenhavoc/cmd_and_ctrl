package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bag End Porter — Creature — Dwarf {3}{G}, 4/4 (The Lord of the Rings:
// Tales of Middle-earth):
//
//	"Whenever this creature attacks, it gets +X/+X until end of turn,
//	 where X is the number of legendary creatures you control."
//
// X is counted as the ability resolves (CR 608.2c), over the legendary
// creatures its controller controls then. The Porter itself is not legendary.
// If the Porter has left the battlefield the pump does nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b80473f7-f1d4-4aa3-988e-1a8610ffbeee",
		Name:         "Bag End Porter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Bag End Porter — it gets +X/+X until end of turn, where X is the number of legendary creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && c.IsLegendary() })
					if x <= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x,
						Label: "Bag End Porter — +X/+X until end of turn"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
