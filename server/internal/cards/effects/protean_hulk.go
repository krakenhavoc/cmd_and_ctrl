package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protean Hulk — Creature — Beast {5}{G}{G}, 6/6 (EDHREC rank 2525):
//
//	"When this creature dies, search your library for any number of
//	 creature cards with total mana value 6 or less, put them onto the
//	 battlefield, then shuffle."
//
// The combo piece that was banned in Commander and then wasn't. A
// dies trigger (graveyard only — an exile or a bounce does not count)
// whose search is the S22 chooser over every creature card in the
// library: the searcher takes as many as they like, the picked set
// is validated against the total-mana-value clause on submit (a
// per-card predicate cannot express it; SearchLibrary.Validate can —
// Myriad Landscape's hook), and an empty pick is the ordinary "fail
// to find". Mana value is printed, X as zero (CR 202.3e). Each card
// enters through the search path, so its own enters-tapped clause
// and every ETB trigger fire, and the library shuffles either way.
//
// No simplification. A fetched permanent whose own entry queues a
// prompt waits for the answer and then arrives (#478), so a Clone
// fetched this way is asked what to copy rather than stranded.
func init() {
	Register(Spec{
		OracleID:     "10180e2f-90c5-4d41-ba44-16b14948f923",
		Name:         "Protean Hulk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Protean Hulk — creature cards with total mana value 6 or less, onto the battlefield", func(g *game.Game, item *game.StackItem) error {
				return b23SearchCreaturesWithTotalManaValue(g, item, 6,
					"Protean Hulk — any number of creature cards with total mana value 6 or less")
			}),
		},
	})
}
