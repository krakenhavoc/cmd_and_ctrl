package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "866586f4-ad97-44fa-a110-6a7a3cb3591b",
		Name:         "Storm Fleet Pyromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventETB, AllOf(Self, tmcB6YouAttackedThisTurn),
					"Storm Fleet Pyromancer — Raid: 2 damage to any target",
					Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
						Instruction(func(ctx *Context) error {
							return sourceDealsDamageToEachLegalTarget(2)(ctx.Game, ctx.Item)
						}),
					}})),
				TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
