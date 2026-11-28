package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud War Beast — Creature — Beast {1}{G}, */* (Exodus, S82 OLD batch 744, #5626):
//
//	"Trample
//	 As this creature enters, choose an opponent.
//	 Skyshroud War Beast's power and toughness are each equal to the number of
//	 nonbasic lands the chosen player controls."
//
// Black Vise's choice, made as it enters (CR 614.12) and stored on the
// permanent, with a characteristic-defining power and toughness (CR 604.3,
// layer 7a) counting the nonbasic lands that player controls as the layers are
// applied. With no opponent chosen it is 0/0. Trample rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6ad434b-6e34-439b-8046-f0047c06e371",
		Name:            "Skyshroud War Beast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AsEnters:        effects.ChoosePlayerAsEnters("Skyshroud War Beast", effects.Opponents),
		Static: []game.StaticAbility{{
			Label:     "Skyshroud War Beast's power and toughness are each equal to the number of nonbasic lands the chosen player controls.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB744NonbasicLandsControlledBy(g, source.ChosenPlayer)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
