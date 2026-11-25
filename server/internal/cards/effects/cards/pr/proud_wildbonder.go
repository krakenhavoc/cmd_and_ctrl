package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proud Wildbonder — Creature — Human Warrior {2}{R/G}{R/G}, 4/3 (Ikoria:
// Lair of Behemoths, #5512):
//
//	"Trample
//	 Creatures you control with trample have \"You may have this creature
//	 assign its combat damage as though it weren't blocked.\""
//
// Siege Behemoth's flag (Characteristic.AssignsAsThoughUnblocked, CR 510.1c)
// on every creature its controller controls that has trample, a layer 6
// static that reads the keyword through the layers. The "may" is the
// engine's per-creature choice as each blocked creature assigns damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e72f995e-288f-4d32-b067-ae73eea87e67",
		Name:            "Proud Wildbonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Creatures you control with trample have \"You may have this creature assign its combat damage as though it weren't blocked.\"",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && game.HasKeyword(target, "trample")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
	})
}
