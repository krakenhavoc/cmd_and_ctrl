package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azorius Charm — Instant {W}{U} (Return to Ravnica, S82 OLD):
//
//	"Choose one —
//	 • Creatures you control gain lifelink until end of turn.
//	 • Draw a card.
//	 • Put target attacking or blocking creature on top of its owner's
//	 library."
//
// Jeskai Charm's modal shape. The first mode grants lifelink to the creatures
// you control as it resolves (CR 611.2c); the third targets an attacking or
// blocking creature and tucks it on top of its owner's library, and does
// nothing if the target has left combat or the battlefield (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7592863-4aee-47b3-9abf-4b863ab0b6ec",
		Name:         "Azorius Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Creatures you control gain lifelink until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GrantKeywordUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"lifelink"},
						Label: "Azorius Charm — lifelink",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerCombatGrant}),
			effects.ModeWithPurpose(effects.ModeDoing("Draw a card.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{N: 1}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerValue, Draws: 1}),
			effects.ModeWithPurpose(effects.ModeDoing("Put target attacking or blocking creature on top of its owner's library.",
				effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return ctx.Game.TuckToLibraryForEffect(id, false)
				})), game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}
