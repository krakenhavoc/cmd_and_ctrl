package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ringsight — Sorcery {1}{U}{B}:
//
//	"The Ring tempts you. Search your library for a card that shares a
//	 color with a legendary creature you control, reveal it, put it
//	 into your hand, then shuffle."
//
// The search runs after the tempt (TheRingTemptsYou.Then), so the
// colors are read from the board the tempt left: "Your Ring-bearer is
// legendary, so you can search for a card that shares a color with the
// Ring-bearer that you chose as the Ring tempted you", and a colorless
// legendary creature adds no color — "Ringsight won't let you search
// for a colorless card" (2023-06-16 rulings). With no colored
// legendary creature the search finds nothing, and the library is
// still searched and shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66e42be3-5125-41b4-9efe-8a6a66899bb6",
		Name:         "Ringsight",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TheRingTemptsYou{Then: ringsightSearch}.Apply(ctx)
		},
	})
}

// ringsightSearch is the sentence after the tempt. The legendary
// creatures are read once, before the search, with the layers caught up
// so the new Ring-bearer's legendary supertype is seen.
func ringsightSearch(ctx *Context, _ uuid.UUID) error {
	legends := legendaryCreaturesYouControl(ctx.Game, ctx.Controller())
	return b06TutorToHand("Ringsight — a card that shares a color with a legendary creature you control", func(c game.Card) bool {
		for _, l := range legends {
			if sharesAColorNow(c, l) {
				return true
			}
		}
		return false
	})(ctx.Item, ctx)
}
