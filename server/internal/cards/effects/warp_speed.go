package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5a824931-2b65-42b1-8653-d4313a6693d2",
		Name:         "Warp Speed",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Replacements: []game.ReplacementEffect{b16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Warp Speed — search for a basic land card or Planet card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return IsBasicLand(c) || c.HasSubtypeWord("Planet") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Warp Speed — a basic land card or Planet card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Strange new worlds — {7}{U}{U}, Sacrifice this enchantment: Draw two cards. This ability costs {X} less to activate, where X is the number of differently named lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    Plus(ManaCost("{7}{U}{U}"), SacrificeThis()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(trekB9DifferentlyNamedPermanentsYouControl(Land()),
					"This ability costs {X} less to activate, where X is the number of differently named lands you control."),
			},
			Effect: Do(DrawCards{N: 2}),
		}},
	})
}
