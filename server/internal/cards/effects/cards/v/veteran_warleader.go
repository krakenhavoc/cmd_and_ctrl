package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Warleader — Creature — Human Soldier Ally {1}{G}{W}, */* (Battle
// for Zendikar, #5547):
//
//	"Veteran Warleader's power and toughness are each equal to the number of
//	 creatures you control.
//	 Tap another untapped Ally you control: This creature gains your choice
//	 of first strike, vigilance, or trample until end of turn."
//
// The size is a characteristic-defining ability (CR 604.3, layer 7a) counting
// itself. The cost taps another untapped Ally (not the {T} symbol, so
// summoning sickness does not matter, CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "784bda96-5bc1-47bc-a69e-e38b047ed914",
		Name:         "Veteran Warleader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Veteran Warleader's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap another untapped Ally you control: This creature gains your choice of first strike, vigilance, or trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapAnotherUntapped("another untapped Ally you control", effects.Creature(), effects.HasSubtype("Ally")),
			Effect:  effects.S82OldB315WarleaderAsk,
		}},
	})
}
