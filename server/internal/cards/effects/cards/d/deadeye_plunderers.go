package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Plunderers — Creature — Human Pirate {3}{U}{B}, 3/3 (Ixalan):
//
//	"This creature gets +1/+1 for each artifact you control.
//	 {2}{U}{B}: Create a Treasure token."
//
// Iron Man Armor's counted layer 7c static (the Plunderers is no artifact, so
// it never counts itself) and Bootleggers' Stash's Treasure ability as a
// stack ability of the permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "597afe7c-a4ba-48d0-9c37-d9affe5ba195",
		Name:         "Deadeye Plunderers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+1 for each artifact you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source != nil && target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.ArtifactsControlledBy(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{B}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
