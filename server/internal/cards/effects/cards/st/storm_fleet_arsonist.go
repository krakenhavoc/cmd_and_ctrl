package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Fleet Arsonist — Creature — Orc Pirate {4}{R}, 4/4 (Ixalan):
//
//	"Raid — When this creature enters, if you attacked this turn, target
//	 opponent sacrifices a permanent of their choice."
//
// Deadeye Tormentor's raid trigger with Misguided Rage's edict: the opponent
// chooses as it resolves, lands included, and the sacrifice does not target
// (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36081fbb-413f-4924-a536-82add34e38be",
		Name:         "Storm Fleet Arsonist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.AvatarB11CondYouAttacked{},
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
					"Storm Fleet Arsonist — raid: target opponent sacrifices a permanent of their choice",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						player, ok := effects.FirstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						g.PlayerSacrificesForEffect(ctx.Source(), player, effects.SacrificeSpec("a permanent", effects.Permanent()),
							"Storm Fleet Arsonist — sacrifice a permanent")
						return nil
					})),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
