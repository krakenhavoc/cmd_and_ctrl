package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kadena, Slinking Sorcerer — Legendary Creature — Snake Wizard Sorcerer
// {1}{B}{G}{U}, 3/3 (Commander 2019, #5517):
//
//	"The first face-down creature spell you cast each turn costs {3} less to
//	 cast.
//	 Whenever a face-down creature you control enters, draw a card."
//
// The discount prices a spell cast face down while this turn's tally of
// face-down creature spells is still 0 (CR 601.2f, the tally is bumped after
// the cost is determined) and reduces generic mana only, so a morph's {3} is
// free. The trigger is Trail of Mystery's: the permanent as it entered is a
// face-down creature whatever it hides.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4da13eb8-ca18-43f1-b54b-d3ddceb980d2",
		Name:         "Kadena, Slinking Sorcerer",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(3, "The first face-down creature spell you cast each turn costs {3} less to cast.",
				effects.YourSpell(), effects.OldB153FirstFaceDownCreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Cmdr2S37FaceDownCreatureEntered, "Kadena, Slinking Sorcerer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
