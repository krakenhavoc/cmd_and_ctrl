package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Self-Assembler — Artifact Creature — Assembly-Worker {5}, 4/4:
//
//	"When this creature enters, you may search your library for an
//	 Assembly-Worker creature card, reveal it, put it into your hand,
//	 then shuffle."
//
// Goblin Matron's search, narrowed to creature cards: "Assembly-Worker"
// is a word of the source's text (CR 612.2), and the search is optional,
// so a library without one is shuffled and nothing more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f36e058-e5fa-48f9-9996-09b77fc193b3",
		Name:         "Self-Assembler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Self-Assembler — search for an Assembly-Worker creature card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Assembly-Worker") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Self-Assembler — an Assembly-Worker creature card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
