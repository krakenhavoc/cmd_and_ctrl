package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behold the Beyond — Sorcery {5}{B}{B}:
//
//	"Discard your hand. Search your library for three cards, put them
//	 into your hand, then shuffle."
//
// The whole hand goes first (every discard is its own event, so payoffs
// queue), then a three-card tutor. A library with fewer than three
// cards yields what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2402ac0-8ab8-413c-bac4-9f4ba2163c58",
		Name:         "Behold the Beyond",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
				return err
			}
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   3,
				Reveal:  false,
				Shuffle: true,
				Reason:  "Behold the Beyond — search your library for three cards",
			}.Apply(ctx)
		},
	})
}
