package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thorin's Last Stand — Instant {2}{W}{W}:
//
//	"Choose one —
//	 • Creatures you control get +2/+1 until end of turn.
//	 • Destroy target artifact or enchantment. You gain 2 life."
//
// Roar of the Kha's anthem bullet (the creatures are fixed as the spell
// resolves, CR 611.2c) and Pyrrhic Strike's destroy bullet. The second
// bullet's only target is the artifact or enchantment, so when it is gone
// the whole mode does nothing and no life is gained (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c4da169-1e50-4790-abdc-214e7b8a2ad5",
		Name:         "Thorin's Last Stand",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Creatures you control get +2/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 1, Label: "Thorin's Last Stand"}.Apply(ctx)
				}),
			ModeDoing("Destroy target artifact or enchantment. You gain 2 life.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				func(item *game.StackItem, ctx *Context, occ int) error {
					if _, ok := ModeTarget(ctx, occ); !ok {
						return nil
					}
					if err := DestroyTheModesTarget(item, ctx, occ); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				}),
		),
	})
}
