package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Fleet Pyromancer — Creature — Human Pirate Wizard {4}{R}, 3/2
// (Ixalan, #5532):
//
//	"Raid — When this creature enters, if you attacked this turn, this creature deals 2 damage to any target."
//
// Gorehorn Raider's body. Raid is an intervening "if" (CR 603.4): the
// trigger goes on the stack only if you attacked this turn, and it
// checks again as it resolves. The damage comes from this creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "866586f4-ad97-44fa-a110-6a7a3cb3591b",
		Name:         "Storm Fleet Pyromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
					"Storm Fleet Pyromancer — Raid: 2 damage to any target",
					effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
						effects.Instruction(func(ctx *effects.Context) error {
							return effects.SourceDealsDamageToEachLegalTarget(2)(ctx.Game, ctx.Item)
						}),
					}})),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
