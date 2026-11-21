package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambitious Farmhand // Seasoned Cathar — Creature — Human Peasant {1}{W},
// 1/1 // Creature — Human Knight 3/3 (Midnight Hunt, Innistrad
// Remastered):
//
//	Ambitious Farmhand
//	  "When this creature enters, you may search your library for a basic
//	   Plains card, reveal it, put it into your hand, then shuffle.
//	   Coven — {1}{W}{W}: Transform this creature. Activate only if you
//	   control three or more creatures with different powers."
//	Seasoned Cathar
//	  "Lifelink"
//
// Coven is read through the layered view as the ability is activated and
// is a restriction on activating, not an intervening if (it is not a
// trigger): trekB48Coven counts the controller's creatures with different
// powers. The activation turns the permanent over in place.
//
// No simplification.
func init() {
	const oracle = "b85e7ab2-6f57-4b65-8c16-034b342996f0"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Ambitious Farmhand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ambitious Farmhand — search for a basic Plains card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandWithSubtype("plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Ambitious Farmhand — a basic Plains card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Coven — {1}{W}{W}: Transform this creature. Activate only if you control three or more creatures with different powers.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{W}{W}"),
			Condition: effects.TrekB48Coven,
			Effect:    effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Seasoned Cathar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
