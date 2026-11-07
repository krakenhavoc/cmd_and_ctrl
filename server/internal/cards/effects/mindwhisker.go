package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindwhisker — Creature — Rat Wizard {2}{U}, 3/2 (Bloomburrow, #4467):
//
//	"At the beginning of your upkeep, surveil 1.
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, creatures your opponents control get -1/-0."
//
// The upkeep surveil is the shared Surveil primitive. The threshold
// half is a layer 7c modify over every creature an opponent controls,
// the shape of Elesh Norn, Grand Cenobite's second static, gated on
// the seven-card count read live on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f823b122-94ac-42ba-b511-9e54d64db112",
		Name:         "Mindwhisker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mindwhisker — surveil 1", Do(Surveil{N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:    "Threshold — Creatures your opponents control get -1/-0 while there are seven or more cards in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller &&
					fraThresholdMet(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
			},
		}},
	})
}
