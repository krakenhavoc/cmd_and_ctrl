package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpse Connoisseur — Creature — Zombie Wizard {4}{B}, 3/3 (Shards of
// Alara, #5588):
//
//	"When this creature enters, you may search your library for a
//	 creature card, put that card into your graveyard, then shuffle.
//	 Unearth {3}{B}"
//
// Vile Entomber's search narrowed to a creature card and made a "you
// may": the searcher can decline the search and the shuffle, and a
// search with a stated quality may find nothing (CR 701.23a). Unearth is
// the shared constructor (CR 702.84a), and an unearthed Connoisseur's
// enters trigger fires as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f39e2611-c837-4ad7-b3a8-1ee48b74777c",
		Name:         "Corpse Connoisseur",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{SelfMillTutor: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Corpse Connoisseur — search for a creature card, put it into your graveyard", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() },
					Dest:      game.ZoneGraveyard,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Corpse Connoisseur — a creature card, into your graveyard",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{Unearth("{3}{B}")},
	})
}
