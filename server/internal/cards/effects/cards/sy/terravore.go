package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terravore — Creature — Lhurgoyf {1}{G}{G}, */* (Dominaria Remastered,
// #5948):
//
//	"Trample
//	 Terravore's power and toughness are each equal to the number of
//	 land cards in all graveyards."
//
// A characteristic-defining ability (CR 604.3, 613.4a): a layer 7a
// self-only static, the same shape as Barrowgoyf's. It counts land cards
// in every graveyard (GapB137LandCardsInAllGraveyards). The layer cache is
// invalidated whenever a card enters or leaves any graveyard, so the size
// is current at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f44fb61a-9aa5-4178-b958-5115d62a7f50",
		Name:            "Terravore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to the number of land cards in all graveyards.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.GapB137LandCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
