package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherwing, Golden-Scale Flagship — Legendary Artifact — Vehicle */4,
// the BACK FACE of Invasion of Kaladesh (face 1):
//
//	"Flying
//	 Aetherwing's power is equal to the number of artifacts you control.
//	 Crew 1"
//
// Registered under "<oracle_id>#1". The power is a layer 7a
// characteristic-defining ability that applies while it is a Vehicle too
// (CR 604.3), and it counts Aetherwing itself, an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        invasionOfKaladeshOracleID + "#1",
		Name:            "Aetherwing, Golden-Scale Flagship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Aetherwing's power is equal to the number of artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = countControlled(g, source.Controller, func(o game.Card) bool { return o.IsArtifact() })
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Aetherwing, Golden-Scale Flagship"),
		}},
	})
}
