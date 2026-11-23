package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunblade Samurai — Enchantment Creature — Human Samurai {4}{W}, 4/4 (Neon
// Dynasty, #5493):
//
//	"Vigilance
//	 Channel — {2}, Discard this card: Search your library for a basic
//	 Plains card, reveal it, put it into your hand, then shuffle. You gain
//	 2 life."
//
// The search is Abzan Monument's, narrowed to Plains; the life is gained
// whether or not the search finds a card, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b767bd0a-a8fe-49d7-a3e7-4f9ac6a3280f",
		Name:            "Sunblade Samurai",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {2}, Discard this card: Search your library for a basic Plains card, reveal it, put it into your hand, then shuffle. You gain 2 life.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandOfAnySubtype("Plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Sunblade Samurai — a basic Plains card",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			},
		}},
	})
}
