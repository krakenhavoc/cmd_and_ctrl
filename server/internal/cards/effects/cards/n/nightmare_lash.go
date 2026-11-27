package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightmare Lash — Artifact — Equipment {4} (Mirrodin, #5606):
//
//	"Equipped creature gets +1/+1 for each Swamp you control.
//	 Equip—Pay 3 life."
//
// Banshee's Blade's counted pump over the Swamps its controller controls, read
// on every layer pass, and Murderer's Axe's equip with a cost that is not mana.
// Equip is sorcery speed as ever (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "414f1c42-6b0e-49ce-8b44-17a75b996b68",
		Name:         "Nightmare Lash",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				swamp := source.TextSubtype("Swamp")
				return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype(swamp) })
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipPayingAbility("Equip—Pay 3 life", effects.PayLife(3))},
	})
}
