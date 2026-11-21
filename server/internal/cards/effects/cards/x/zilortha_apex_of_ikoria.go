package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zilortha, Apex of Ikoria — Legendary Creature — Dinosaur 8/8, the back
// face of Invasion of Ikoria (March of the Machine, S82 batch 23):
//
//	"Reach
//	 For each non-Human creature you control, you may have that creature
//	 assign its combat damage as though it weren't blocked."
//
// Imaryll's option (Characteristic.AssignsAsThoughUnblocked) set on every
// non-Human creature its controller controls, Zilortha included (she is a
// Dinosaur). It is a layer 6 static read through the layers, so a creature
// that stops being a creature, becomes a Human, or changes controller
// stops having it, and each blocked creature's controller is asked
// separately whether to use it (CR 510.1c). "Human" is a word of the
// source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        effects.InvasionOfIkoriaOracleID + "#1",
		Name:            "Zilortha, Apex of Ikoria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "For each non-Human creature you control, you may have that creature assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() &&
					!target.HasSubtype(source.TextSubtype("Human"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
	})
}
