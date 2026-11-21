package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaos Wand — Artifact {3} (Outlaws of Thunder Junction Commander):
//
//	"{4}, {T}: Target opponent exiles cards from the top of their library
//	 until they exile an instant or sorcery card. You may cast that card
//	 without paying its mana cost. Then put the exiled cards that weren't
//	 cast this way on the bottom of that library in a random order."
//
// Dazzling Sphinx's walk (Game.ExileUntilCastFreeByForEffect): the opponent
// exiles, the Wand's controller is offered the free cast during the
// resolution, and everything not cast goes under its owner's library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b95c115-6745-4ce1-896b-9021e93c161e",
		Name:         "Chaos Wand",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Target opponent exiles cards from the top of their library until an instant or sorcery card; you may cast it without paying its mana cost",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Cmdr2S43ExileUntilInstantOrSorcery(g, item, effects.TargetedPlayerOf(effects.NewContext(g, item)))
			},
		}},
	})
}
