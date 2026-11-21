package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Druid — Creature — Human Druid {1}{G}, 0/3 (Eldritch Moon, #5541):
//
//	"When this creature dies, you may search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Farhaven Elf's fetch on a dies trigger. "You may search" is the
// search's own Optional (#3053), so the player is asked as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d031d7d2-1e41-4201-9073-c4c1be182ba4",
		Name:         "Primal Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Primal Druid — you may search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Optional:      true,
					Reason:        "Primal Druid — a basic land card, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
