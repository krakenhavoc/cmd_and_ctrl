package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twitch — Instant {2}{U} (Worldwake, S82 OLD b488):
//
//	"You may tap or untap target artifact, creature, or land.
//	 Draw a card."
//
// Captain of the Mists' tap-or-untap pick, made as the spell resolves, then the draw as a
// separate step (CR 608.2c): the draw waits for the pick's answer and is made whatever
// it was (the "may" can be declined entirely). A target that has become illegal still
// leaves the spell resolving for its draw only when it had no other target to lose: with
// its single target gone the spell is countered on resolution (CR 608.2b), so no card is
// drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "321fd970-a259-448d-bd98-e85e386174f9",
		Name:         "Twitch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B38TapOrUntapTarget("Twitch")(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
