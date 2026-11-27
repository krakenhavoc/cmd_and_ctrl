package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// March of the Machines — {3}{U} Enchantment (Archenemy):
//
//	"Each noncreature artifact is an artifact creature with power and
//	 toughness each equal to its mana value. (Equipment that's a creature
//	 can't equip a creature.)"
//
// Two statics over every noncreature artifact on the battlefield, whoever
// controls it: the creature type in layer 4, then base power and toughness
// equal to the artifact's own mana value in layer 7b (Starfield of Nyx's
// shape, which reads the mana value off the affected card). The Equipment
// rider is the rules' own (CR 301.5c), not the card's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51092634-308e-4779-aa51-182715dbc734",
		Name:         "March of the Machines",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Each noncreature artifact is an artifact creature.",
				Layer:     game.Layer4Type,
				AppliesTo: marchOfTheMachinesApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			{
				Label:     "Each such artifact has power and toughness each equal to its mana value.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: marchOfTheMachinesAppliesToPT,
				Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
					mv := target.ManaValue()
					c.Power, c.Toughness = mv, mv
				},
			},
		},
	})
}

func marchOfTheMachinesApplies(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsArtifact() && !target.IsCreature()
}

// marchOfTheMachinesAppliesToPT is the same set read in layer 7, after layer 4
// has made each of them a creature: "noncreature" is judged by the printed
// card, so the artifacts this static animated keep matching.
func marchOfTheMachinesAppliesToPT(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsArtifact() && !target.PrintedIsCreature()
}
