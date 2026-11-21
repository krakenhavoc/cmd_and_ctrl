package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warp Speed — Enchantment {2}{G} (Star Trek, #4973):
//
//	"Lands you control enter untapped.
//	 When this enchantment enters, search your library for a basic land
//	 card or Planet card, reveal it, put it into your hand, then shuffle.
//	 Strange new worlds — {7}{U}{U}, Sacrifice this enchantment: Draw two
//	 cards. This ability costs {X} less to activate, where X is the number
//	 of differently named lands you control."
//
// The first line is Horizon Explorer's replacement
// (b16LandsYouControlEnterUntapped, CR 614.12). The search admits a
// basic land card or any card with the Planet subtype, and may find
// nothing. "Strange new worlds" is an ability word with no rules meaning;
// the reduction is Suus Mahna Practitioner's (generic mana only,
// CR 118.7a), counted as the cost is determined (CR 602.2b, 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a824931-2b65-42b1-8653-d4313a6693d2",
		Name:         "Warp Speed",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Replacements: []game.ReplacementEffect{effects.B16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Warp Speed — search for a basic land card or Planet card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtypeWord("Planet") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Warp Speed — a basic land card or Planet card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Strange new worlds — {7}{U}{U}, Sacrifice this enchantment: Draw two cards. This ability costs {X} less to activate, where X is the number of differently named lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    effects.Plus(effects.ManaCost("{7}{U}{U}"), effects.SacrificeThis()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.TrekB9DifferentlyNamedPermanentsYouControl(effects.Land()),
					"This ability costs {X} less to activate, where X is the number of differently named lands you control."),
			},
			Effect: effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
