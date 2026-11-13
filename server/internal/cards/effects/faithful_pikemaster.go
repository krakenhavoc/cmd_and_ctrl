package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithful Pikemaster — Creature — Rhino Monk Soldier {3}{W}, 3/4
// (Foundations Jumpstart, #5464):
//
//	"When this creature enters, scry 2.
//	 As long as it's your turn, this creature has first strike."
//
// The enters trigger is the shared scry 2 (Archive Dragon). The first
// strike is Skilled Battlecarver's layer-6 static: it reads whose turn
// it is each time the layers are applied, so the Pikemaster has first
// strike on its controller's turn and not on anyone else's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17d5c53d-7c77-45c2-b8be-7e43c1ecd902",
		Name:         "Faithful Pikemaster",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faithful Pikemaster — scry 2", Do(Scry{N: 2})),
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "first strike")
			},
		}},
	})
}
