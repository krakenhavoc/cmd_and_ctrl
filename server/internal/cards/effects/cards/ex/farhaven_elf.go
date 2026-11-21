package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farhaven Elf — Creature — Elf Druid {2}{G}, 1/1 (EDHREC rank 1485):
//
//	"When this creature enters, you may search your library for a
//	 basic land card, put it onto the battlefield tapped, then
//	 shuffle."
//
// Solemn Simulacrum's front half on an Elf. The "you may" is the
// optional-trigger prompt; the search is the searcher's choice of
// basic, arriving tapped, then the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ce2357f-93e6-40ca-beca-8f4e15adc464",
		Name:         "Farhaven Elf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Farhaven Elf — search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Farhaven Elf — a basic land, tapped",
				}.Apply(effects.NewContext(g, item))
			}), "Farhaven Elf — search for a basic land?"),
		},
	})
}
