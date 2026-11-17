package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Welkin Hawk — Creature — Bird {1}{W}, 1/1 (Exodus, #5626):
//
//	"Flying
//	 When this creature dies, you may search your library for a card
//	 named Welkin Hawk, reveal that card, put it into your hand, then
//	 shuffle."
//
// Flying is a printed keyword. The dies trigger is Tempest Hawk's
// search on WhenThisDies, optional (CR 603.5): accepted, it searches for
// a card named Welkin Hawk, reveals it, puts it into your hand and
// shuffles, and you may fail to find one (CR 701.23b). It triggers
// only when the Hawk goes from the battlefield to a graveyard
// (CR 700.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4547625a-0caf-4ecc-a65f-0fd2cad70294",
		Name:            "Welkin Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisDies("Welkin Hawk — search your library for a card named Welkin Hawk",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Welkin Hawk") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Welkin Hawk — a card named Welkin Hawk",
					}.Apply(NewContext(g, item))
				}),
				"Search your library for a card named Welkin Hawk?"),
		},
	})
}
