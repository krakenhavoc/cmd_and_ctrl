package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enigma Drake — Creature — Drake {1}{U}{R}, */4 (Foundations, #4465):
//
//	"Flying
//	 Enigma Drake's power is equal to the number of instant and sorcery
//	 cards in your graveyard."
//
// A characteristic-defining ability (CR 604.3), applied in layer 7a as
// Melek, Reforged Researcher's is, with the same count. It sets the power
// only; the toughness stays the printed 4. The count is read live, so it
// follows the graveyard as cards enter and leave it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62007675-8a0d-4211-83b2-0daf3641dedc",
		Name:            "Enigma Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Enigma Drake's power is equal to the number of instant and sorcery cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.FandanielInstantsAndSorceries(g, source.Controller)
			},
		}},
	})
}
