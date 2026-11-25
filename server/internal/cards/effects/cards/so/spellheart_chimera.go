package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellheart Chimera — Creature — Chimera {1}{U}{R}, */3 (Commander 2016, S82
// older sets, #5537):
//
//	"Flying, trample
//	 Spellheart Chimera's power is equal to the number of instant and sorcery
//	 cards in your graveyard."
//
// A characteristic-defining ability (CR 604.3, 613.4a): the power is set in
// the layer-7a bucket and follows the graveyard live, whatever else is
// layered over it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9745fc92-4955-48e3-a6c8-1300891cf078",
		Name:            "Spellheart Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Static: []game.StaticAbility{{
			Label:    "Spellheart Chimera's power is equal to the number of instant and sorcery cards in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.CmdrB17InstantSorceryCardsInGraveyard(g, source.Controller)
			},
		}},
	})
}
