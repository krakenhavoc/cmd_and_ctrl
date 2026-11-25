package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Outburst — Sorcery {3}{R}{R} (Dominaria, S82 OLD, #5530):
//
//	"Chandra's Outburst deals 4 damage to target player or planeswalker.
//	 Search your library and/or graveyard for a card named Chandra, Bold
//	 Pyromancer, reveal it, and put it into your hand. If you search your
//	 library this way, shuffle."
//
// The search is its own sentence and not a "you may", so it runs whether or
// not the damage did anything. The engine asks which zones to search, shows
// and shuffles the library only if it was chosen (the planeswalker-fetch
// cycle's SearchLibrary with AlsoGraveyard). A target that became illegal
// fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "894a344e-a8a7-4ddc-833d-eea41f1918d7",
		Name:         "Chandra's Outburst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Tutors: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 4))},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DamageToFirstTargetStep(4)(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				const named = "Chandra, Bold Pyromancer"
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     func(c game.Card) bool { return c.Name == named },
					Dest:          game.ZoneHand,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					AlsoGraveyard: true,
					Reason:        "Choose a card named " + named + " to put into your hand",
				}.Apply(ctx)
			}),
		),
	})
}
