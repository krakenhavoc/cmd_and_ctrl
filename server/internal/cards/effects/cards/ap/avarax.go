package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avarax — Creature — Beast {3}{R}{R}, 3/3 (Onslaught):
//
//	"Haste
//	 When this creature enters, you may search your library for a card
//	 named Avarax, reveal it, put it into your hand, then shuffle.
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// Tempest Hawk's search for its own name, as an optional enters trigger
// (CR 603.5): declining searches nothing and does not shuffle. In a
// singleton Commander deck there is no second Avarax to find, but the
// search and the shuffle still happen when the player agrees. The pump
// is Furnace Gremlin's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e1deb7e-d40d-437e-b817-b66132d2d1d6",
		Name:            "Avarax",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Avarax — search your library for a card named Avarax",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Avarax") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Avarax — a card named Avarax",
					}.Apply(effects.NewContext(g, item))
				}),
				"Search your library for a card named Avarax?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Avarax — +1/+0 until end of turn"),
		}},
	})
}
