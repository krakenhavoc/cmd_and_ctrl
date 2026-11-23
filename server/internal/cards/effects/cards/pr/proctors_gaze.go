package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a7fe7621-9bb1-4d0d-8b50-c179c685d79d",
		Name:         "Proctor's Gaze",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceChosenTarget(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
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
