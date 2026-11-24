package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geist-Honored Monk — Creature — Human Monk {3}{W}{W}, */* (Kaldheim
// Commander, #5506):
//
//	"Vigilance
//	 Geist-Honored Monk's power and toughness are each equal to the number
//	 of creatures you control.
//	 When this creature enters, create two 1/1 white Spirit creature
//	 tokens with flying."
//
// Crusader of Odric's characteristic-defining ability (layer 7a), read
// live, the Monk itself counted, so it is a 3/3 once its Spirits land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5deac0b-4417-42cc-b145-db0afe34f6e0",
		Name:            "Geist-Honored Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Geist-Honored Monk's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Geist-Honored Monk — create two 1/1 white Spirit tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 2})),
		},
	})
}
