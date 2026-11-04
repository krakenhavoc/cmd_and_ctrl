package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Proctor's Gaze — Instant {2}{G}{U} (Secrets of Strixhaven, #4456):
//
//	"Return up to one target nonland permanent to its owner's hand. Search
//	 your library for a basic land card, put it onto the battlefield tapped,
//	 then shuffle."
//
// Two instructions in printed order (CR 608.2c): the bounce, then Rampant
// Growth's search. The bounce is a step of its own, so a commander the owner is
// asked about (CR 903.9) finishes before the search starts. With no target
// chosen the spell still searches; if the one target chosen is illegal on
// resolution the spell does nothing (CR 608.2b), the search included. The
// controller picks which basic land, and a library with none shuffles anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7fe7621-9bb1-4d0d-8b50-c179c685d79d",
		Name:         "Proctor's Gaze",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return bounceChosenTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Proctor's Gaze — a basic land",
				}.Apply(ctx)
			}),
		),
	})
}
