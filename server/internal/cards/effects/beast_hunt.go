package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast Hunt — Sorcery {3}{G} (Planechase, #5584):
//
//	"Reveal the top three cards of your library. Put all creature cards
//	 revealed this way into your hand and the rest into your graveyard."
//
// Mulch's body over creatures and three cards: "all creature cards" is not
// a choice, so there is no prompt, and the rest are put into the graveyard,
// not milled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "069c8fa1-fa04-4620-9356-a618158dad3e",
		Name:         "Beast Hunt",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cmdr2B15RevealTopTakeAllMatching(ctx.Game, item.Controller, ctx.Source(), 3,
				"Beast Hunt — reveal the top three cards, creatures to your hand and the rest to your graveyard", Creature())
		},
	})
}
