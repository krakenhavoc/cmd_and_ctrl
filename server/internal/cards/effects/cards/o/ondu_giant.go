package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ondu Giant — {3}{G} Creature — Giant Druid 2/4 (Jumpstart 2022, #5948):
//
//	"When this creature enters, you may search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Rampart Architect's optional basic-land search on an enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed8b2887-15fa-4e62-83ac-55973ac6f76c",
		Name:         "Ondu Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Ondu Giant — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Optional:      true,
						Shuffle:       true,
						Source:        item.SourceCardID,
						Reason:        "Ondu Giant — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}), game.Purpose{Lands: 1}),
		},
	})
}
