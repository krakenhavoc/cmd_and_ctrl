package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simulacrum Shaper — Creature — Elf Druid {1}{G}{G}, 2/2:
//
//	"When this creature enters, you may search your library for a basic
//	 land card, put that card onto the battlefield tapped, then shuffle.
//	 When this creature dies, draw a card."
//
// The "you may" is the trigger's own prompt (CR 603.5); the search then
// fetches one basic land tapped (the search's own TappedOnEntry flag,
// which runs the entry replacement pipeline) and shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e45e2da-7064-4ec0-8d09-55fc7d1aeaa8",
		Name:         "Simulacrum Shaper",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Simulacrum Shaper — search for a basic land, put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Simulacrum Shaper — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}), "Simulacrum Shaper — search your library for a basic land card?"),
			effects.WhenThisDies("Simulacrum Shaper — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
