package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorehorn Raider — Creature — Minotaur Pirate {4}{R}, 4/4 (Foundations,
// #4465):
//
//	"Raid — When this creature enters, if you attacked this turn, this
//	 creature deals 2 damage to any target."
//
// Fire Nation Raider's raid enters trigger with a target. Raid is an
// intervening "if" (CR 603.4): checked as it triggers and again as it
// resolves, and any attack declared this turn counts. The target is chosen
// as the trigger goes on the stack, and the Minotaur deals the damage, as it
// last existed if it has left (sourceDealsDamageToEachLegalTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41699d92-c209-4fcb-bbd6-685b3437d0c2",
		Name:         "Gorehorn Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
					"Gorehorn Raider — Raid: 2 damage to any target",
					effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
						effects.Instruction(func(ctx *effects.Context) error {
							return effects.SourceDealsDamageToEachLegalTarget(2)(ctx.Game, ctx.Item)
						}),
					}})),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
