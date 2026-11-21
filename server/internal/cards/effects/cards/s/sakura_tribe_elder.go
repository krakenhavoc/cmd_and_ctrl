package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sakura-Tribe Elder — Creature — Snake Shaman {1}{G}, 1/1:
//
//	"Sacrifice Sakura-Tribe Elder: Search your library for a basic
//	land card, put it onto the battlefield tapped, then shuffle."
//
// The classic "chump block, then ramp anyway" body — the ability has
// no mana or tap component, so it can be activated at any time
// including after blocking, which is the whole trick.
//
// No tap cost means summoning sickness never applies: the Elder can
// sacrifice itself the turn it lands.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3afc704-220f-498f-9eaa-0821b17dc24c",
		Name:         "Sakura-Tribe Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice Sakura-Tribe Elder: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Sakura-Tribe Elder — a basic land",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
