package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oriq Loremage — Creature — Human Warlock {2}{B}{B}, 3/3 (Strixhaven:
// School of Mages, #5502):
//
//	"{T}: Search your library for a card, put it into your graveyard, then
//	 shuffle. If it's an instant or sorcery card, put a +1/+1 counter on
//	 this creature."
//
// The search is Goblin Engineer's into the graveyard; the counter rides the
// search's keyed tail, read off the card that was put there, and goes on
// only while the Loremage is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d1dbe80-9ad0-41ad-8ba6-e1094f596eaf",
		Name:         "Oriq Loremage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Search your library for a card, put it into your graveyard, then shuffle. If it's an instant or sorcery card, put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:   item.Controller,
					Dest:     game.ZoneGraveyard,
					Limit:    1,
					Shuffle:  true,
					Reason:   "Oriq Loremage — a card, into your graveyard",
					Source:   item.SourceCardID,
					ThenTail: game.KeyedTail{Ref: effects.S82OldB40LoremageCounter, Player: item.Controller, Source: item.SourceCardID},
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
