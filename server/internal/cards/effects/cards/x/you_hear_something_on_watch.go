package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "55d94c09-b6aa-4af5-923a-b4efc2a245df",
		Name:         "You Hear Something on Watch",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Rouse the Party — Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "You Hear Something on Watch — Rouse the Party"}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			effects.ModeWithPurpose(effects.ModeDoing("Set Off Traps — This spell deals 5 damage to target attacking creature.",
				effects.TargetCreature("target attacking creature", effects.B41Attacking()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})), effects.ForTargets(effects.DamageToTarget(0, 5))),
		),
	})
}
