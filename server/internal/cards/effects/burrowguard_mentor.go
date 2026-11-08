package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrowguard Mentor — Creature — Rabbit Soldier {G}{W}, */*
// (Bloomburrow, #4467):
//
//	"Trample
//	 Burrowguard Mentor's power and toughness are each equal to the
//	 number of creatures you control."
//
// Adeline's layer 7a characteristic-defining ability (CR 604.3, 613.4a)
// on both numbers: self-only, set on every recompute to the count of
// creatures its controller controls (b04CreaturesControlled), the Mentor
// itself included. Trample rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f293da27-e42d-4e87-8b61-d3c27e4c6d71",
		Name:            "Burrowguard Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Burrowguard Mentor's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
