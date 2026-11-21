package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perilous Forays — Enchantment {3}{G}{G} (Ravnica Remastered):
//
//	"{1}, Sacrifice a creature: Search your library for a land card with a
//	 basic land type, put it onto the battlefield tapped, then shuffle."
//
// A land TYPE, not the basic supertype (hasABasicLandType), as Boseiju does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d552f5cb-19b6-4def-90d5-6d2d9570f93d",
		Name:         "Perilous Forays",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Search your library for a land card with a basic land type, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Source:        item.SourceCardID,
					Predicate:     hasABasicLandType,
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Perilous Forays — a land card with a basic land type",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
