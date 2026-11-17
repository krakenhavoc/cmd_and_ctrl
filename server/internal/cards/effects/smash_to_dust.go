package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smash to Dust — Sorcery {1}{R} (Dominaria United, #5487):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target creature with defender.
//	 • Smash to Dust deals 1 damage to each creature your opponents
//	   control."
//
// The first two bullets target and the engine applies the chosen bullet's
// own clause, so the picker offers artifacts or defenders depending on the
// mode. "Defender" is read through the layered view, so a creature that
// gained or lost it counts as it is. The third bullet has no target and is
// Boiling Earth's damage to each creature an opponent controls, dealt as
// one instance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c13955b8-bad3-4f79-ad9f-11e214a1cab2",
		Name:         "Smash to Dust",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
				game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(Mode("Destroy target creature with defender.",
				TargetCreature("target creature with defender", HasKeyword("defender"))),
				game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(Mode("Smash to Dust deals 1 damage to each creature your opponents control."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(2) {
				return damageEachMatching(ctx, And(Creature(), OpponentControls()), 1)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
