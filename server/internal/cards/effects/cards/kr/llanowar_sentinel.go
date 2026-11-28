package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Llanowar Sentinel — Creature — Elf {2}{G}, 2/3 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"When this creature enters, you may pay {1}{G}. If you do, search your library for a card
//	 named Llanowar Sentinel, put that card onto the battlefield, then shuffle."
//
// Subway Train's optional payment: {1}{G} is paid as the trigger resolves, and only if it was does
// the search run, over cards named Llanowar Sentinel, to the battlefield, with the shuffle that ends it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "272ab5d5-7edb-4457-8a09-cabe2325c881",
		Name:         "Llanowar Sentinel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Llanowar Sentinel — you may pay {1}{G} to search for a card named Llanowar Sentinel", llanowarSentinelOffer),
		},
	})
}

func llanowarSentinelOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{G}",
		Question: "Llanowar Sentinel — pay {1}{G} to search your library for a card named Llanowar Sentinel?",
		Then:     llanowarSentinelPaid,
	}.Apply(effects.NewContext(g, item))
}

var llanowarSentinelPaid = effects.PayUnlessBranches("pay-unless/s82-old-b777-llanowar-sentinel",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.SearchLibrary{
			Player:    r.Chooser,
			Predicate: func(c game.Card) bool { return c.Name == "Llanowar Sentinel" },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Reason:    "Llanowar Sentinel — a card named Llanowar Sentinel, put onto the battlefield",
		}.Apply(ctx)
	}, nil)
