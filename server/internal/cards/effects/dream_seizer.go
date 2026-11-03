package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Seizer — Creature — Faerie Rogue {3}{B}, 3/2:
//
//	"Flying
//	 When this creature enters, you may blight 1. If you do, each opponent
//	 discards a card. (To blight 1, put a -1/-1 counter on a creature you
//	 control.)"
//
// The blight is asked as the trigger resolves (Blight, #4683); "if you
// do" is a creature chosen, and then each opponent discards a card of
// their choice from their own hand (eachOpponentDiscardsOne). Flying is
// printed card data.
func init() {
	Register(Spec{
		OracleID:     "c7e4bc5d-e1d9-4fbc-8d53-0cec73086828",
		Name:         "Dream Seizer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dream Seizer — you may blight 1; if you do, each opponent discards a card", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 1, May: true, Then: dreamSeizerDiscard}.Apply(NewContext(g, item))
			}),
		},
	})
}

// dreamSeizerDiscard is "if you do, each opponent discards a card". An
// on-disk key.
var dreamSeizerDiscard = RegisterBlighted("blight/dream-seizer-discard", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return eachOpponentDiscardsOne(ctx.Game, ctx.Item)
})
