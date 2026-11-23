package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Militant Inquisitor — Creature — Human Cleric {2}{W}, 2/3 (Foundations
// Jumpstart):
//
//	"This creature gets +1/+0 for each Equipment you control."
//
// A layer 7c self-modify whose count is re-read on every layer pass
// (Adelbert Steiner's shape): every Equipment its controller controls
// counts, attached to it or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7",
		Name:         "Militant Inquisitor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each Equipment you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Equipment") })
			},
		}},
	})
}
