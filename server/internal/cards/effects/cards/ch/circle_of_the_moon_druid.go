package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circle of the Moon Druid — Creature — Human Elf Druid {2}{G}, 2/4
// (Adventures in the Forgotten Realms):
//
//	"Bear Form — During your turn, this creature is a Bear with base power
//	 and toughness 4/2. (It loses all other creature types.)"
//
// One printed ability, two effects while it is its controller's turn:
// Goddric's shape. A layer 4 replacement of its creature types with Bear
// (card types stay) and a layer 7b base 4/2. Counters and pumps still apply
// on top of the base (CR 613.4c).
//
// No simplification.
func init() {
	const label = "Bear Form — During your turn, this creature is a Bear with base power and toughness 4/2."
	duringYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "90569b99-8b63-4d2e-b7de-b578b8adbdcb",
		Name:         "Circle of the Moon Druid",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     label,
				Layer:     game.Layer4Type,
				AppliesTo: duringYourTurn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.SetCreatureTypes(source.TextSubtypes("Bear"))
				},
			},
			{
				Label:     label,
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: duringYourTurn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 4, 2
				},
			},
		},
	})
}
