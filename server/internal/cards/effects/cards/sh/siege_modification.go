package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siege Modification — {1}{R}{R} Enchantment — Aura (Aether Revolt, #5536):
//
//	"Enchant creature or Vehicle
//	 As long as enchanted permanent is a Vehicle, it's a creature in addition
//	 to its other types.
//	 Enchanted creature gets +3/+0 and has first strike."
//
// The Vehicle clause is a layer 4 static, so the host is a creature by the
// time the layer 6 and 7 clauses read it (CR 613.1): an uncrewed Vehicle gets
// the bonus too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac6cf9a7-732d-4ff0-9e88-aae97a2ff169",
		Name:         "Siege Modification",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Or(effects.Creature(), effects.Subtype("Vehicle"))),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			{
				Label: "As long as enchanted permanent is a Vehicle, it's a creature in addition to its other types.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && target.HasSubtypeWord("Vehicle")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("first strike"),
		},
	})
}
