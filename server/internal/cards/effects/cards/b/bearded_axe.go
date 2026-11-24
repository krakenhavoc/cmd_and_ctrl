package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bearded Axe — Artifact — Equipment {2}{R} (Kaldheim, #5505):
//
//	"Equipped creature gets +1/+1 for each Dwarf, Equipment, and/or Vehicle
//	 you control.
//	 Equip {2}"
//
// "And/or" counts each permanent once however many of the three it is, and
// the Axe counts itself. "You" is the Axe's controller, read on every
// recompute through the layered view (a changeling is a Dwarf).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a231319d-5d72-4aa6-8868-5987694791cd",
		Name:         "Bearded Axe",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				dwarf := source.TextSubtype("Dwarf")
				return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool {
					return c.HasSubtype(dwarf) || c.HasSubtype("Equipment") || c.HasSubtype("Vehicle")
				})
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
