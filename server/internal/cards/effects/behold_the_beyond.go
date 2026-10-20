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
		// The search waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				_, err := discardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:  ctx.Controller(),
					Dest:    game.ZoneHand,
					Limit:   3,
					Reveal:  false,
					Shuffle: true,
					Reason:  "Behold the Beyond — search your library for three cards",
				}.Apply(ctx)
			}),
		),
	})
}
