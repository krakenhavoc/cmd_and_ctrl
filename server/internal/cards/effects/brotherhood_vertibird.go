package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Vertibird — Artifact — Vehicle {3}, */4 (Fallout, #5948):
//
//	"Flying
//	 Brotherhood Vertibird's power is equal to the number of artifacts you
//	 control.
//	 Crew 2"
//
// The power is a characteristic-defining ability (CR 604.3), a layer 7a
// self-only static (Dodgy Jalopy's shape). A Vehicle is an artifact, so it
// counts itself, crewed or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b000071-3768-44cf-8957-180d0e94c988",
		Name:            "Brotherhood Vertibird",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of artifacts you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = countControlled(g, source.Controller, func(p game.Card) bool { return p.IsArtifact() })
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Brotherhood Vertibird"),
		}},
	})
}
