package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vesuva — Land (EDHREC top 1,000, #3654):
//
//	"You may have this land enter tapped as a copy of any land on the
//	 battlefield."
//
// Clone's enters-as-a-copy replacement over lands, with the tapped rider
// as the except clause: it enters tapped only when the copy is taken. A
// player who declines gets Vesuva untapped, with no abilities, as
// printed; the copy keeps no memory of being Vesuva.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4001b868-ada1-43f4-92e2-27ab0e80c913",
		Name:         "Vesuva",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Vesuva",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsLand() })
				},
				func(ev *game.ReplacementEvent, _ *game.PrintedValues, _ *game.Game, _ *game.Card) {
					ev.EntersTapped = true
				},
			),
		},
	})
}
