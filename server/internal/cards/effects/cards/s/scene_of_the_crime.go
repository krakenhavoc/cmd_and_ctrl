package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scene of the Crime — Artifact Land — Clue (EDHREC rank 2588):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}, Tap an untapped creature you control: Add one mana of any
//	 color.
//	 {2}, Sacrifice this land: Draw a card."
//
// The land that is a Clue. The tapped entry is the real CR 614
// self-replacement; the colourless tap is an ordinary mana ability;
// the crack is the Clue token's own activated ability — "{2},
// Sacrifice this: draw a card", no tap in the cost, so a land that
// entered tapped can still be cracked at once — declared as a CR 602
// ability with a mana-and-sacrifice-self cost.
//
// #758 supplies the coloured mana ability's second cost component.
// The creature tap is not the {T} symbol, so a creature that entered
// this turn may pay it; the land itself still pays the printed {T}.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba11a517-1dbd-4797-9f5e-46ce0f6c77c0",
		Name:         "Scene of the Crime",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{
					Tap: true,
					TapOthers: &game.TapOthersCost{
						Count:  1,
						Filter: effects.TargetPermanent("an untapped creature you control", effects.Creature()),
						Label:  "an untapped creature you control",
					},
				},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
