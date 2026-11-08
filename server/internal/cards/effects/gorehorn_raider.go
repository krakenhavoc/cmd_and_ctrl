package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "41699d92-c209-4fcb-bbd6-685b3437d0c2",
		Name:         "Gorehorn Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventETB, AllOf(Self, tmcB6YouAttackedThisTurn),
					"Gorehorn Raider — Raid: 2 damage to any target",
					Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
						Instruction(func(ctx *Context) error {
							return sourceDealsDamageToEachLegalTarget(2)(ctx.Game, ctx.Item)
						}),
					}})),
				TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
