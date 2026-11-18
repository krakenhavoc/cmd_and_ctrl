package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You Hear Something on Watch — Instant {1}{W} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Choose one —
//	 • Rouse the Party — Creatures you control get +1/+1 until end of turn.
//	 • Set Off Traps — This spell deals 5 damage to target attacking creature."
//
// The pump is Ambitious Assault's mass boost: the set is the creatures you
// control as the spell resolves, locked then (CR 611.2c). The damage is
// dealt by the spell to an attacking creature, which must still be attacking
// as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55d94c09-b6aa-4af5-923a-b4efc2a245df",
		Name:         "You Hear Something on Watch",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Rouse the Party — Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "You Hear Something on Watch — Rouse the Party"}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			ModeWithPurpose(ModeDoing("Set Off Traps — This spell deals 5 damage to target attacking creature.",
				TargetCreature("target attacking creature", b41Attacking()),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})), ForTargets(DamageToTarget(0, 5))),
		),
	})
}
