package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyxathid — Creature — Elemental {1}{B}{B}, 7/7 (Jumpstart, #5510):
//
//	"As this creature enters, choose an opponent.
//	 This creature gets -1/-1 for each card in the chosen player's hand."
//
// Entropic Specter's as-enters choice, stored on the permanent
// (ChoosePlayerAsEnters). The shrink is a layer 7c modification that reads
// the chosen player's hand live, with the hand-size hint so the layer cache
// is dropped whenever a hand changes. Until the choice is made it applies
// to nothing, so the 7/7 is not read off nobody's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c889b2e-380a-4bc4-8d56-5fb12bc7c78b",
		Name:         "Nyxathid",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePlayerAsEnters("Nyxathid", effects.Opponents),
		Static: []game.StaticAbility{{
			Label:             "This creature gets -1/-1 for each card in the chosen player's hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.ChosenPlayerOf(g, source.InstanceID) != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CardsInHand(g, effects.ChosenPlayerOf(g, source.InstanceID))
				c.Power -= n
				c.Toughness -= n
			},
		}},
	})
}
