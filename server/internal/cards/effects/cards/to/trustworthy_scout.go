package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trustworthy Scout — Creature — Human Scout {1}{W}, 2/2 (Modern Horizons):
//
//	"{1}{W}, Exile this card from your graveyard: Search your library for a
//	 card named Trustworthy Scout, reveal it, put it into your hand, then
//	 shuffle."
//
// An activated ability that functions only from its owner's graveyard
// (CR 113.6b), Renew's shape without the sorcery restriction: the cost
// moves the card to exile before the ability is on the stack. The search
// may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51d7b287-90d5-4e3b-a82c-6b98996154e1",
		Name:         "Trustworthy Scout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Exile this card from your graveyard: Search your library for a card named Trustworthy Scout, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.Name == "Trustworthy Scout" },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Trustworthy Scout — a card named Trustworthy Scout",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
