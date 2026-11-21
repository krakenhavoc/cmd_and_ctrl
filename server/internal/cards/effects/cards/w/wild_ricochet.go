package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Ricochet — Instant {2}{R}{R} (Starter Commander Decks):
//
//	"You may choose new targets for target instant or sorcery spell. Then
//	 copy that spell. You may choose new targets for the copy."
//
// Two steps in printed order, the second waiting for the first's prompt (CR
// 608.2c): the optional choose-new-targets offer on the spell (CR 115.7c),
// then Increasing Vengeance's copy, which the caster controls (CR 707.10)
// and may retarget. Any instant or sorcery spell is a legal target, whoever
// controls it; a spell with no targets cannot be retargeted but is still
// copied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c35fd11-be45-4984-bd83-6e4f3fbc47a9",
		Name:         "Wild Ricochet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				id := ts[0].ID
				return effects.ChangeTargets{
					StackID:  id,
					Policy:   game.RetargetChooseNew,
					Optional: true,
					Reason:   "Wild Ricochet — you may choose new targets for the spell",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				id := ts[0].ID
				return effects.CopySpell{
					StackID:          id,
					Controller:       ctx.Controller(),
					Count:            1,
					ChooseNewTargets: true,
				}.Apply(ctx)
			}),
		),
	})
}
