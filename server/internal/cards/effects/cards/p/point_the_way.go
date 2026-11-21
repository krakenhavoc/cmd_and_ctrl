package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Point the Way — Enchantment {G}:
//
//	"Start your engines!
//	 {3}{G}, Sacrifice this enchantment: Search your library for up to X
//	 basic land cards, where X is your speed. Put those cards onto the
//	 battlefield tapped, then shuffle."
//
// ADR 0138 (#2122, #2711). X is read as the ability resolves (CR 608.2h)
// from its controller's speed, so it counts a speed gained in response; a
// player with no speed searches for nothing and still shuffles. The
// sacrifice is the cost, so the speed the enchantment started stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ffe9c8b-e0cb-4a29-bc14-649dbec5d1f3",
		Name:            "Point the Way",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, Sacrifice this enchantment: Search your library for up to X basic land cards, where X is your speed. Put those cards onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.YourSpeed(g, item.Controller)
				if x <= 0 {
					return g.ShuffleLibraryForEffect(item.Controller)
				}
				return effects.SearchLibrary{
					Player:        item.Controller,
					Source:        item.SourceCardID,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         x,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Point the Way — basic lands onto the battlefield tapped",
				}.Apply(ctx)
			},
		}},
	})
}
