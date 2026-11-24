package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raven Wings — Artifact — Equipment {2} (Kaldheim, #5505):
//
//	"Equipped creature gets +1/+0, has flying, and is a Bird in addition to
//	 its other types.
//	 Equip {2}"
//
// The Bird type is a layer-4 addition (CR 613.1d), so the creature keeps
// every type it had; the subtype is a word of the card's text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d668f67-7cde-447a-8eb3-35c0c5c5ab69",
		Name:         "Raven Wings",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
			{
				Label:     "Equipped creature is a Bird in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					bird := source.TextSubtype("Bird")
					for _, s := range c.Subtypes {
						if s == bird {
							return
						}
					}
					c.Subtypes = append(append([]string(nil), c.Subtypes...), bird)
				},
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
