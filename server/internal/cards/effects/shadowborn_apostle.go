package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadowborn Apostle — Creature — Human Cleric {B}, 1/1:
//
//	"A deck can have any number of cards named Shadowborn Apostle.
//	 {B}, Sacrifice six creatures named Shadowborn Apostle: Search your
//	 library for a Demon creature card, put it onto the battlefield,
//	 then shuffle."
//
// The deck rule is CR 113.6n's, read off the oracle text by
// internal/deck for every card, catalog or not. The activation's cost
// is six creatures named Shadowborn Apostle, the source included, all
// leaving as one simultaneous exit (SacrificeN). The name is the
// permanent's name as it is now, so a token copy of an Apostle counts
// and an Apostle turned into a copy of something else does not. The
// search has a stated quality, so the searcher may find nothing
// (CR 701.23b), and the library is shuffled either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "595a15f0-77f3-4544-8acc-10630e12cc14",
		Name:         "Shadowborn Apostle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice six creatures named Shadowborn Apostle: Search your library for a Demon creature card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{B}"), SacrificeN(6, "six creatures named Shadowborn Apostle",
				Creature(), shadowbornApostleNamed)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtype("Demon") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Shadowborn Apostle — search for a Demon creature card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// shadowbornApostleNamed is "named Shadowborn Apostle".
func shadowbornApostleNamed(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return game.HasName(c, "Shadowborn Apostle")
}
