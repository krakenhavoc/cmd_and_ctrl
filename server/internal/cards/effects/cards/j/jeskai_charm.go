package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeskai Charm — Instant {U}{R}{W} (S82 GAP, #5948):
//
//	"Choose one —
//	 • Put target creature on top of its owner's library.
//	 • Jeskai Charm deals 4 damage to target opponent or planeswalker.
//	 • Creatures you control get +1/+1 and gain lifelink until end of turn."
//
// Banishment Decree's tuck, Barbed Lightning's modal damage, and an Indelicate
// Healer-style lifelink grant. The creatures you control as the third mode
// resolves are the ones pumped (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a9a05b4-1910-4f97-a0fe-0f57017f6df0",
		Name:         "Jeskai Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Put target creature on top of its owner's library.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return ctx.Game.TuckToLibraryForEffect(id, false)
				})), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Jeskai Charm deals 4 damage to target opponent or planeswalker.",
				effects.TargetOpponentOrPlaneswalker(),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 4}.Apply(ctx)
				})), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures you control get +1/+1 and gain lifelink until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
						Label: "Jeskai Charm — +1/+1",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"lifelink"},
						Label: "Jeskai Charm — lifelink",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
		),
	})
}
