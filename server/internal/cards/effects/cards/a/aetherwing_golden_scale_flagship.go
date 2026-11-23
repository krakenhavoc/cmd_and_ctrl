package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        effects.InvasionOfKaladeshOracleID + "#1",
		Name:            "Aetherwing, Golden-Scale Flagship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Aetherwing's power is equal to the number of artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.CountControlled(g, source.Controller, func(o game.Card) bool { return o.IsArtifact() })
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Aetherwing, Golden-Scale Flagship"),
		}},
	})
}
