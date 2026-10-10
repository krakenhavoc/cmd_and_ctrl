package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surgical Precision — Sorcery {1}{W} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Destroy target creature with toughness 4 or greater. You gain 1 life.
//	 • You draw a card and gain 2 life."
//
// The life is gained whether or not the creature was destroyed (an
// indestructible one survives), as printed; a target that has gone makes
// the bullet do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0caccec3-4942-499a-9434-0c94389afcf8",
		Name:         "Surgical Precision",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target creature with toughness 4 or greater. You gain 1 life.",
				TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				}),
			ModeDoing("You draw a card and gain 2 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				}),
		),
	})
}
